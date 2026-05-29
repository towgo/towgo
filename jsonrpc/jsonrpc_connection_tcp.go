package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net"
	"sync"
	"time"
)

const (
	// BUFFERLENGTH is the default maximum TCP read buffer size.
	BUFFERLENGTH int64 = 1024000
	// DATAEND marks the end of one framed TCP JSON-RPC message.
	DATAEND byte = '\n'
)

// TcpRpcConnection adapts a TCP connection to JSON-RPC.
type TcpRpcConnection struct {
	guid                   string
	isConnected            bool
	remoteAddr             string
	rpcRequest             *Jsonrpcrequest
	rpcResponse            *Jsonrpcresponse
	requestCallBackFuncs   sync.Map
	requestCallBackCancels sync.Map
	conn                   net.Conn
	bodyStr                string
	paramsBytes            []byte
	resultBytes            []byte
	ctx                    context.Context
	nextFunc               func()
	err                    error
	result                 any
	values                 sync.Map
}

// NewTcpRpcConnection creates a TCP JSON-RPC connection from a framed message.
func NewTcpRpcConnection(conn net.Conn, bodyStr string) *TcpRpcConnection {
	if bodyStr == "" {
		return nil
	}
	rpcRequest, _ := ToJsonrpcrequest(bodyStr)
	rpcResponse := NewJsonrpcresponse()
	if rpcRequest.Method == "" {
		response, _ := ToJsonrpcresponse(bodyStr)
		rpcResponse = &response
	}
	return &TcpRpcConnection{
		guid:        newConnectionGUID("TCP:"),
		isConnected: true,
		rpcRequest:  rpcRequest,
		rpcResponse: rpcResponse,
		conn:        conn,
		bodyStr:     bodyStr,
	}
}

// AnalysisByString parses a raw JSON-RPC message into request or response state.
func (c *TcpRpcConnection) AnalysisByString(message string) {
	c.bodyStr = message
	rpcRequest, _ := ToJsonrpcrequest(message)
	c.rpcRequest = rpcRequest
	c.paramsBytes = nil
	if rpcRequest.Method == "" {
		response, _ := ToJsonrpcresponse(message)
		c.rpcResponse = &response
	} else {
		c.rpcResponse = NewJsonrpcresponse()
	}
}

// GetRemoteAddr returns the TCP peer address.
func (c *TcpRpcConnection) GetRemoteAddr() string {
	if c.remoteAddr != "" {
		return c.remoteAddr
	}
	if c.conn != nil && c.conn.RemoteAddr() != nil {
		c.remoteAddr = c.conn.RemoteAddr().String()
	}
	return c.remoteAddr
}

// Read returns the raw JSON-RPC message body.
func (c *TcpRpcConnection) Read() string {
	return c.bodyStr
}

// ReadParams unmarshals request params into each destination.
func (c *TcpRpcConnection) ReadParams(destParams ...interface{}) error {
	if len(c.paramsBytes) == 0 {
		var err error
		c.paramsBytes, err = json.Marshal(c.rpcRequest.Params)
		if err != nil {
			return err
		}
	}
	for _, dest := range destParams {
		if err := json.Unmarshal(c.paramsBytes, dest); err != nil {
			return err
		}
	}
	return nil
}

// ReadResult unmarshals response result into each destination.
func (c *TcpRpcConnection) ReadResult(destResult ...interface{}) error {
	if len(c.resultBytes) == 0 {
		var err error
		c.resultBytes, err = json.Marshal(c.rpcResponse.Result)
		if err != nil {
			return err
		}
	}
	for _, dest := range destResult {
		if err := json.Unmarshal(c.resultBytes, dest); err != nil {
			return err
		}
	}
	return nil
}

// SetResult stores a response result without writing immediately.
func (c *TcpRpcConnection) SetResult(result interface{}) {
	c.result = result
}

// GetResult returns the stored response result.
func (c *TcpRpcConnection) GetResult() interface{} {
	return c.result
}

// WriteResult stores a result and writes the response frame.
func (c *TcpRpcConnection) WriteResult(result interface{}) {
	c.result = result
	c.Write()
}

// Write writes the JSON-RPC response frame to the TCP connection.
func (c *TcpRpcConnection) Write() {
	if c.rpcResponse == nil {
		c.rpcResponse = NewJsonrpcresponse()
	}
	c.rpcResponse.Id = c.rpcRequest.Id
	c.rpcResponse.Timestampin = c.rpcRequest.Timestampin
	c.rpcResponse.Timestampout = timestampMillis()
	if c.err != nil {
		c.rpcResponse.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, c.err.Error())
	} else if c.result != nil {
		c.rpcResponse.Result = c.result
	}
	if c.conn == nil {
		return
	}
	data, _ := json.Marshal(c.rpcResponse)
	var buffer bytes.Buffer
	buffer.Write(data)
	buffer.WriteByte(DATAEND)
	_, _ = c.conn.Write(buffer.Bytes())
}

// GetRpcRequest returns the request payload.
func (c *TcpRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.rpcRequest
}

// GetRpcResponse returns the response payload.
func (c *TcpRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.rpcResponse
}

// Push sends a one-way JSON-RPC request frame.
func (c *TcpRpcConnection) Push(request *Jsonrpcrequest) error {
	if request == nil || request.Method == "" {
		return errors.New("method not set")
	}
	if request.Id == "" {
		request.Id = randomString(64)
	}
	request.Timestampin = timestampMillis()
	return c.writeFrame(request)
}

// Call sends a JSON-RPC request frame and stores a callback for the response.
func (c *TcpRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if rpcRequest == nil {
		rpcRequest = NewJsonrpcrequest()
	}
	rpcRequest.Id = newConnectionGUID("")
	rpcRequest.Jsonrpc = "2.0"
	rpcRequest.Timestampin = timestampMillis()
	if callback != nil {
		c.requestCallBackFuncs.Store(rpcRequest.Id, callback)
		ctx, cancel := context.WithCancel(context.Background())
		c.requestCallBackCancels.Store(rpcRequest.Id, cancel)
		go c.clearCallback(ctx, rpcRequest.Id)
	}
	_ = c.writeFrame(rpcRequest)
}

// writeFrame marshals a value and writes one newline-delimited TCP frame.
func (c *TcpRpcConnection) writeFrame(v any) error {
	if c.conn == nil {
		return errors.New("tcp connection is not available")
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.Write(data)
	buffer.WriteByte(DATAEND)
	_, err = c.conn.Write(buffer.Bytes())
	return err
}

// clearCallback removes a pending callback after cancellation or timeout.
func (c *TcpRpcConnection) clearCallback(ctx context.Context, requestID string) {
	timer := time.NewTimer(600 * time.Second)
	select {
	case <-ctx.Done():
		timer.Stop()
	case <-timer.C:
	}
	c.requestCallBackFuncs.Delete(requestID)
	c.requestCallBackCancels.Delete(requestID)
}

// LinkType returns tcp.
func (c *TcpRpcConnection) LinkType() string {
	return "tcp"
}

// IsConnected reports whether the TCP connection is active.
func (c *TcpRpcConnection) IsConnected() bool {
	return c.isConnected
}

// GUID returns the connection identifier.
func (c *TcpRpcConnection) GUID() string {
	return c.guid
}

// EnableHealthCheck is reserved for TCP health checks.
func (c *TcpRpcConnection) EnableHealthCheck() {
}

// DisableHealthCheck is reserved for TCP health checks.
func (c *TcpRpcConnection) DisableHealthCheck() {
}

// WriteError writes an error response frame.
func (c *TcpRpcConnection) WriteError(code int64, msg string) {
	c.rpcResponse.Error.Set(code, msg)
	c.Write()
}

// WriteResponse writes a complete response frame.
func (c *TcpRpcConnection) WriteResponse(resp Jsonrpcresponse) {
	c.rpcResponse = &resp
	c.Write()
}

// Close closes the TCP connection.
func (c *TcpRpcConnection) Close() {
	c.isConnected = false
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// SetValue stores a connection-scoped value.
func (c *TcpRpcConnection) SetValue(key string, value any) {
	c.values.Store(key, value)
}

// GetValue loads a connection-scoped value.
func (c *TcpRpcConnection) GetValue(key string) (value any, ok bool) {
	return c.values.Load(key)
}

// Context returns the runtime context attached to the connection.
func (c *TcpRpcConnection) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// WithContext replaces the runtime context attached to the connection.
func (c *TcpRpcConnection) WithContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
}

// GetError returns the stored execution error.
func (c *TcpRpcConnection) GetError() error {
	return c.err
}

// WithError stores an execution error.
func (c *TcpRpcConnection) WithError(err error) {
	c.err = err
}

// Next continues the middleware chain.
func (c *TcpRpcConnection) Next() {
	if c.nextFunc != nil {
		c.nextFunc()
	}
}

// SetNextFunc sets the continuation used by Next.
func (c *TcpRpcConnection) SetNextFunc(fn func()) {
	c.nextFunc = fn
}

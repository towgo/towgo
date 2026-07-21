package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
)

// HttpRpcConnection adapts one HTTP request and response writer to JSON-RPC.
type HttpRpcConnection struct {
	guid        string
	isConnected bool
	written     bool
	remoteAddr  string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	response    http.ResponseWriter
	request     *http.Request
	bodyStr     string
	paramsBytes []byte
	resultBytes []byte
	result      any
	err         error
	ctx         context.Context
	values      sync.Map
}

// NewHttpRpcConnection creates an HTTP JSON-RPC connection from net/http values.
func NewHttpRpcConnection(w http.ResponseWriter, r *http.Request) *HttpRpcConnection {
	conn := &HttpRpcConnection{
		guid:        newConnectionGUID("HTTP:"),
		response:    w,
		request:     r,
		rpcResponse: NewJsonrpcresponse(),
		isConnected: true,
	}
	bodyStr := conn.Read()
	if bodyStr == "" {
		return nil
	}
	rpcRequest, err := ToJsonrpcrequest(bodyStr)
	if err != nil {
		conn.rpcResponse.Error.Set(JSONRPC_400_BAD_REQUEST, err.Error())
	}
	conn.rpcRequest = rpcRequest
	conn.rpcResponse.Id = rpcRequest.Id
	return conn
}

// GetRemoteAddr returns the client IP address for the HTTP request.
func (c *HttpRpcConnection) GetRemoteAddr() string {
	if c.remoteAddr != "" {
		return c.remoteAddr
	}
	c.remoteAddr = remoteIP(c.request)
	return c.remoteAddr
}

// Read returns the cached HTTP request body.
func (c *HttpRpcConnection) Read() string {
	if c.bodyStr != "" {
		return c.bodyStr
	}
	if c.request == nil || c.request.Body == nil {
		return ""
	}
	data, _ := io.ReadAll(c.request.Body)
	c.bodyStr = string(data)
	return c.bodyStr
}

// ReadParams unmarshals request params into each destination.
func (c *HttpRpcConnection) ReadParams(destParams ...interface{}) error {
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
func (c *HttpRpcConnection) ReadResult(destResult ...interface{}) error {
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
func (c *HttpRpcConnection) SetResult(result interface{}) {
	c.result = result
}

// GetResult returns the stored response result.
func (c *HttpRpcConnection) GetResult() interface{} {
	return c.result
}

// WriteResult stores a result and writes the response.
func (c *HttpRpcConnection) WriteResult(result interface{}) {
	c.result = result
	c.Write()
}

// Write writes the JSON-RPC response to the HTTP response writer.
func (c *HttpRpcConnection) Write() {
	if c.response == nil {
		return
	}
	c.written = true
	c.rpcResponse.Id = c.rpcRequest.Id
	c.rpcResponse.Timestampin = c.rpcRequest.Timestampin
	c.rpcResponse.Timestampout = timestampMillis()
	if c.err != nil {
		c.rpcResponse.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, c.err.Error())
	} else if c.result != nil {
		c.rpcResponse.Result = c.result
	}
	_ = json.NewEncoder(c.response).Encode(c.rpcResponse)
}

// GetRpcRequest returns the request payload.
func (c *HttpRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.rpcRequest
}

// GetRpcResponse returns the response payload.
func (c *HttpRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.rpcResponse
}

// Push writes a JSON-RPC request to the HTTP response writer.
func (c *HttpRpcConnection) Push(request *Jsonrpcrequest) error {
	if request == nil || request.Method == "" {
		return errors.New("method not set")
	}
	if request.Id == "" {
		request.Id = randomString(64)
	}
	request.Timestampin = timestampMillis()
	return json.NewEncoder(c.response).Encode(request)
}

// Call reports that HTTP does not support full-duplex JSON-RPC calls.
func (c *HttpRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if callback == nil {
		return
	}
	c.WithError(errors.New("http protocol does not support call"))
	callback(c)
}

// LinkType returns http.
func (c *HttpRpcConnection) LinkType() string {
	return "http"
}

// IsConnected reports whether the HTTP connection is active.
func (c *HttpRpcConnection) IsConnected() bool {
	return c.isConnected
}

// GUID returns the connection identifier.
func (c *HttpRpcConnection) GUID() string {
	return c.guid
}

// EnableHealthCheck is a no-op for HTTP connections.
func (c *HttpRpcConnection) EnableHealthCheck() {
}

// DisableHealthCheck is a no-op for HTTP connections.
func (c *HttpRpcConnection) DisableHealthCheck() {
}

// WriteError writes an error response.
func (c *HttpRpcConnection) WriteError(code int64, msg string) {
	c.rpcResponse.Error.Set(code, msg)
	c.Write()
}

// WriteResponse writes a complete response payload.
func (c *HttpRpcConnection) WriteResponse(resp Jsonrpcresponse) {
	c.rpcResponse = &resp
	c.Write()
}

// Close marks the HTTP connection as closed.
func (c *HttpRpcConnection) Close() {
	c.isConnected = false
}

// SetValue stores a connection-scoped value.
func (c *HttpRpcConnection) SetValue(key string, value any) {
	c.values.Store(key, value)
}

// GetValue loads a connection-scoped value.
func (c *HttpRpcConnection) GetValue(key string) (value any, ok bool) {
	return c.values.Load(key)
}

// Context returns the runtime context attached to the connection.
func (c *HttpRpcConnection) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	if c.request != nil {
		return c.request.Context()
	}
	return context.Background()
}

// WithContext replaces the runtime context attached to the connection.
func (c *HttpRpcConnection) WithContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
}

// GetError returns the stored execution error.
func (c *HttpRpcConnection) GetError() error {
	return c.err
}

// WithError stores an execution error.
func (c *HttpRpcConnection) WithError(err error) {
	c.err = err
}

// Next continues the middleware chain.
func (c *HttpRpcConnection) Next() {
	if fn, ok := c.GetValue("nextFunc"); ok {
		if next, ok := fn.(func()); ok {
			next()
		}
	}
}

// SetNextFunc sets the continuation used by Next.
func (c *HttpRpcConnection) SetNextFunc(fn func()) {
	c.SetValue("nextFunc", fn)
}

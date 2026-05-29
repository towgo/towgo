package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// TOWGO_WEBSOCKET_PING is the reserved method name for websocket ping messages.
const TOWGO_WEBSOCKET_PING = "towgo.websocket.ping"

// clientCallTimeOut is the default websocket client call timeout in seconds.
var clientCallTimeOut int64 = 600

// WebSocketRpcConnection adapts a websocket connection to JSON-RPC.
type WebSocketRpcConnection struct {
	lock        *sync.Mutex
	guid        string
	isConnected bool
	messageChan chan string
	// CallTimeOut is the timeout in seconds for pending websocket calls.
	CallTimeOut            int64
	requestBody            string
	request                *Jsonrpcrequest
	response               *Jsonrpcresponse
	wsConn                 *websocket.Conn
	paramsBytes            []byte
	resultBytes            []byte
	rpcCallBackFuncs       *sync.Map
	requestCallBackCancels *sync.Map
	ctx                    context.Context
	nextFunc               func()
	err                    error
	result                 any
	values                 *sync.Map
}

// RpcCallback stores a pending websocket request callback.
type RpcCallback struct {
	// Request is the outbound request waiting for a response.
	Request *Jsonrpcrequest
	// Callback is invoked when the matching response is received or times out.
	Callback func(JsonRpcConnection)
}

// SetClientCallTimeOut sets the websocket client call timeout in seconds.
func SetClientCallTimeOut(second int64) {
	clientCallTimeOut = second
}

// NewWebSocketRpcConnection creates a websocket JSON-RPC connection.
func NewWebSocketRpcConnection(ws *websocket.Conn) *WebSocketRpcConnection {
	return &WebSocketRpcConnection{
		lock:                   &sync.Mutex{},
		guid:                   newConnectionGUID("WS:"),
		isConnected:            ws != nil,
		messageChan:            make(chan string, 1),
		CallTimeOut:            clientCallTimeOut,
		wsConn:                 ws,
		response:               NewJsonrpcresponse(),
		rpcCallBackFuncs:       &sync.Map{},
		requestCallBackCancels: &sync.Map{},
		values:                 &sync.Map{},
	}
}

// Duplicate creates a lightweight view sharing the same websocket state.
func (c *WebSocketRpcConnection) Duplicate() *WebSocketRpcConnection {
	next := NewWebSocketRpcConnection(c.wsConn)
	next.guid = c.guid
	next.lock = c.lock
	next.isConnected = c.isConnected
	next.messageChan = c.messageChan
	next.CallTimeOut = c.CallTimeOut
	next.rpcCallBackFuncs = c.rpcCallBackFuncs
	next.requestCallBackCancels = c.requestCallBackCancels
	next.values = c.values
	return next
}

// AnalysisByString parses a raw JSON-RPC message into request or response state.
func (c *WebSocketRpcConnection) AnalysisByString(message string) {
	c.requestBody = message
	rpcRequest, _ := ToJsonrpcrequest(message)
	c.request = rpcRequest
	c.paramsBytes = nil
	if rpcRequest.Method == "" {
		response, _ := ToJsonrpcresponse(message)
		c.response = &response
	} else {
		c.response = NewJsonrpcresponse()
	}
}

// GetRemoteAddr returns the websocket peer address.
func (c *WebSocketRpcConnection) GetRemoteAddr() string {
	if c.request != nil && c.request.Route.SourceAddr != "" {
		return c.request.Route.SourceAddr
	}
	if c.wsConn != nil && c.wsConn.RemoteAddr() != nil {
		return c.wsConn.RemoteAddr().String()
	}
	return ""
}

// Read returns the raw JSON-RPC message body.
func (c *WebSocketRpcConnection) Read() string {
	return c.requestBody
}

// ReadParams unmarshals request params into each destination.
func (c *WebSocketRpcConnection) ReadParams(destParams ...interface{}) error {
	if len(c.paramsBytes) == 0 {
		var err error
		c.paramsBytes, err = json.Marshal(c.request.Params)
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
func (c *WebSocketRpcConnection) ReadResult(destResult ...interface{}) error {
	if len(c.resultBytes) == 0 {
		var err error
		c.resultBytes, err = json.Marshal(c.response.Result)
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
func (c *WebSocketRpcConnection) SetResult(result interface{}) {
	c.result = result
}

// GetResult returns the stored response result.
func (c *WebSocketRpcConnection) GetResult() interface{} {
	return c.result
}

// WriteResult stores a result and writes the response message.
func (c *WebSocketRpcConnection) WriteResult(result interface{}) {
	c.result = result
	c.Write()
}

// Write sends the JSON-RPC response through the websocket.
func (c *WebSocketRpcConnection) Write() {
	if c.response == nil {
		c.response = NewJsonrpcresponse()
	}
	if c.request != nil {
		c.response.Id = c.request.Id
		c.response.Timestampin = c.request.Timestampin
	}
	c.response.Timestampout = timestampMillis()
	if c.err != nil {
		c.response.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, c.err.Error())
	} else if c.result != nil {
		c.response.Result = c.result
	}
	if c.wsConn == nil {
		return
	}
	data, _ := json.Marshal(c.response)
	_ = websocket.Message.Send(c.wsConn, string(data))
}

// GetRpcRequest returns the request payload.
func (c *WebSocketRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.request
}

// GetRpcResponse returns the response payload.
func (c *WebSocketRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.response
}

// Push sends a one-way JSON-RPC request through the websocket.
func (c *WebSocketRpcConnection) Push(request *Jsonrpcrequest) error {
	if request == nil || request.Method == "" {
		return errors.New("method not set")
	}
	if c.wsConn == nil {
		return errors.New("websocket connection is not available")
	}
	if request.Id == "" {
		request.Id = randomString(64)
	}
	request.Timestampin = timestampMillis()
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	return websocket.Message.Send(c.wsConn, string(data))
}

// Call sends a JSON-RPC request and stores a callback for the response.
func (c *WebSocketRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if rpcRequest == nil {
		rpcRequest = NewJsonrpcrequest()
	}
	rpcRequest.Id = newConnectionGUID("")
	rpcRequest.Jsonrpc = "2.0"
	rpcRequest.Timestampin = timestampMillis()
	if callback != nil {
		c.rpcCallBackFuncs.Store(rpcRequest.Id, &RpcCallback{Request: rpcRequest, Callback: callback})
		ctx, cancel := context.WithCancel(context.Background())
		c.requestCallBackCancels.Store(rpcRequest.Id, cancel)
		go c.requestDoneOrTimeOut(ctx, rpcRequest.Id)
	}
	if c.wsConn == nil {
		if callback != nil {
			resp := NewJsonrpcresponse()
			resp.Id = rpcRequest.Id
			resp.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, "websocket connection is not available")
			c.response = resp
			callback(c)
		}
		return
	}
	data, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}
	if err := websocket.Message.Send(c.wsConn, string(data)); err != nil && callback != nil {
		resp := NewJsonrpcresponse()
		resp.Id = rpcRequest.Id
		resp.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		c.response = resp
		callback(c)
	}
}

// requestDoneOrTimeOut removes a pending callback or injects a timeout response.
func (c *WebSocketRpcConnection) requestDoneOrTimeOut(ctx context.Context, requestID string) {
	timer := time.NewTimer(time.Duration(c.CallTimeOut) * time.Second)
	select {
	case <-ctx.Done():
		timer.Stop()
	case <-timer.C:
		resp := NewJsonrpcresponse()
		resp.Id = requestID
		resp.Error.Set(JSONRPC_408_REQUEST_TIMEOUT, "REQUEST_TIMEOUT")
		c.MockResponse(resp)
	}
	c.rpcCallBackFuncs.Delete(requestID)
	c.requestCallBackCancels.Delete(requestID)
}

// LinkType returns websocket.
func (c *WebSocketRpcConnection) LinkType() string {
	return "websocket"
}

// MockResponse injects a response into the local receive channel.
func (c *WebSocketRpcConnection) MockResponse(response *Jsonrpcresponse) {
	if response == nil {
		return
	}
	data, _ := json.Marshal(response)
	c.messageChan <- string(data)
}

// ReceiveMessage receives one raw message from the local channel.
func (c *WebSocketRpcConnection) ReceiveMessage() (string, bool) {
	msg, ok := <-c.messageChan
	return msg, ok
}

// IsConnected reports whether the websocket is active.
func (c *WebSocketRpcConnection) IsConnected() bool {
	return c.isConnected
}

// GUID returns the connection identifier.
func (c *WebSocketRpcConnection) GUID() string {
	return c.guid
}

// EnableHealthCheck is reserved for websocket health checks.
func (c *WebSocketRpcConnection) EnableHealthCheck() {
}

// DisableHealthCheck is reserved for websocket health checks.
func (c *WebSocketRpcConnection) DisableHealthCheck() {
}

// WriteError writes an error response message.
func (c *WebSocketRpcConnection) WriteError(code int64, msg string) {
	c.response.Error.Set(code, msg)
	c.Write()
}

// WriteResponse writes a complete response message.
func (c *WebSocketRpcConnection) WriteResponse(resp Jsonrpcresponse) {
	c.response = &resp
	c.Write()
}

// Close closes the websocket connection.
func (c *WebSocketRpcConnection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.isConnected {
		return
	}
	c.isConnected = false
	close(c.messageChan)
	if c.wsConn != nil {
		_ = c.wsConn.Close()
	}
}

// SetValue stores a connection-scoped value.
func (c *WebSocketRpcConnection) SetValue(key string, value any) {
	c.values.Store(key, value)
}

// GetValue loads a connection-scoped value.
func (c *WebSocketRpcConnection) GetValue(key string) (value any, ok bool) {
	return c.values.Load(key)
}

// Context returns the runtime context attached to the connection.
func (c *WebSocketRpcConnection) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// WithContext replaces the runtime context attached to the connection.
func (c *WebSocketRpcConnection) WithContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
}

// GetError returns the stored execution error.
func (c *WebSocketRpcConnection) GetError() error {
	return c.err
}

// WithError stores an execution error.
func (c *WebSocketRpcConnection) WithError(err error) {
	c.err = err
}

// Next continues the middleware chain.
func (c *WebSocketRpcConnection) Next() {
	if c.nextFunc != nil {
		c.nextFunc()
	}
}

// SetNextFunc sets the continuation used by Next.
func (c *WebSocketRpcConnection) SetNextFunc(fn func()) {
	c.nextFunc = fn
}

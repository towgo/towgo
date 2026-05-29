package jsonrpc

import (
	"context"
	"encoding/json"
	"sync"
)

// LocalRpcConnection is an in-memory connection used for direct execution.
type LocalRpcConnection struct {
	guid        string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	ctx         context.Context
	nextFunc    func()
	err         error
	result      any
	values      sync.Map
}

// NewLocalRpcConnection creates an in-memory JSON-RPC connection.
func NewLocalRpcConnection(rpcRequest *Jsonrpcrequest, rpcResponse *Jsonrpcresponse) JsonRpcConnection {
	if rpcRequest == nil {
		rpcRequest = NewJsonrpcrequest()
	}
	if rpcResponse == nil {
		rpcResponse = NewJsonrpcresponse()
	}
	return &LocalRpcConnection{
		guid:        newConnectionGUID("LOCAL:"),
		rpcRequest:  rpcRequest,
		rpcResponse: rpcResponse,
	}
}

// GetRemoteAddr returns the fixed local address label.
func (c *LocalRpcConnection) GetRemoteAddr() string {
	return "local"
}

// Read returns the JSON encoding of the request payload.
func (c *LocalRpcConnection) Read() string {
	data, _ := json.Marshal(c.rpcRequest)
	return string(data)
}

// ReadParams unmarshals request params into each destination.
func (c *LocalRpcConnection) ReadParams(destParams ...interface{}) error {
	data, err := json.Marshal(c.rpcRequest.Params)
	if err != nil {
		return err
	}
	for _, dest := range destParams {
		if err := json.Unmarshal(data, dest); err != nil {
			return err
		}
	}
	return nil
}

// ReadResult unmarshals response result into each destination.
func (c *LocalRpcConnection) ReadResult(destResult ...interface{}) error {
	data, err := json.Marshal(c.rpcResponse.Result)
	if err != nil {
		return err
	}
	for _, dest := range destResult {
		if err := json.Unmarshal(data, dest); err != nil {
			return err
		}
	}
	return nil
}

// SetResult stores a response result without writing immediately.
func (c *LocalRpcConnection) SetResult(result interface{}) {
	c.result = result
}

// GetResult returns the stored response result.
func (c *LocalRpcConnection) GetResult() interface{} {
	return c.result
}

// WriteResult stores a result and builds the response.
func (c *LocalRpcConnection) WriteResult(result interface{}) {
	c.result = result
	c.Write()
}

// Write builds the response payload in memory.
func (c *LocalRpcConnection) Write() {
	c.rpcResponse.Id = c.rpcRequest.Id
	c.rpcResponse.Timestampin = c.rpcRequest.Timestampin
	c.rpcResponse.Timestampout = timestampMillis()
	if c.err != nil {
		c.rpcResponse.Error.Set(JSONRPC_500_INTERNAL_SERVER_ERROR, c.err.Error())
		return
	}
	if c.result != nil {
		c.rpcResponse.Result = c.result
	}
}

// GetRpcRequest returns the request payload.
func (c *LocalRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.rpcRequest
}

// GetRpcResponse returns the response payload.
func (c *LocalRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.rpcResponse
}

// Push is a no-op for local connections.
func (c *LocalRpcConnection) Push(request *Jsonrpcrequest) error {
	return nil
}

// Call invokes the callback with a not-implemented response.
func (c *LocalRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if rpcRequest == nil {
		rpcRequest = NewJsonrpcrequest()
	}
	rpcRequest.Id = newConnectionGUID("")
	rpcRequest.Jsonrpc = "2.0"
	rpcRequest.Timestampin = timestampMillis()
	resp := NewJsonrpcresponse()
	resp.Id = rpcRequest.Id
	resp.Error.Set(JSONRPC_501_NOT_IMPLEMENTED, "local call requires a JsonrpcServer")
	conn := NewLocalRpcConnection(rpcRequest, resp)
	if callback != nil {
		callback(conn)
	}
}

// LinkType returns local.
func (c *LocalRpcConnection) LinkType() string {
	return "local"
}

// IsConnected always reports true for local connections.
func (c *LocalRpcConnection) IsConnected() bool {
	return true
}

// GUID returns the connection identifier.
func (c *LocalRpcConnection) GUID() string {
	return c.guid
}

// EnableHealthCheck is a no-op for local connections.
func (c *LocalRpcConnection) EnableHealthCheck() {
}

// DisableHealthCheck is a no-op for local connections.
func (c *LocalRpcConnection) DisableHealthCheck() {
}

// WriteError stores an error and builds the response.
func (c *LocalRpcConnection) WriteError(code int64, msg string) {
	c.rpcResponse.Error.Set(code, msg)
	c.Write()
}

// WriteResponse replaces the response payload.
func (c *LocalRpcConnection) WriteResponse(resp Jsonrpcresponse) {
	c.rpcResponse = &resp
}

// Close is a no-op for local connections.
func (c *LocalRpcConnection) Close() {
}

// SetValue stores a connection-scoped value.
func (c *LocalRpcConnection) SetValue(key string, value any) {
	c.values.Store(key, value)
}

// GetValue loads a connection-scoped value.
func (c *LocalRpcConnection) GetValue(key string) (value any, ok bool) {
	return c.values.Load(key)
}

// Context returns the runtime context attached to the connection.
func (c *LocalRpcConnection) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// WithContext replaces the runtime context attached to the connection.
func (c *LocalRpcConnection) WithContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.ctx = ctx
}

// GetError returns the stored execution error.
func (c *LocalRpcConnection) GetError() error {
	return c.err
}

// WithError stores an execution error.
func (c *LocalRpcConnection) WithError(err error) {
	c.err = err
}

// Next continues the middleware chain.
func (c *LocalRpcConnection) Next() {
	if c.nextFunc != nil {
		c.nextFunc()
	}
}

// SetNextFunc sets the continuation used by Next.
func (c *LocalRpcConnection) SetNextFunc(fn func()) {
	c.nextFunc = fn
}

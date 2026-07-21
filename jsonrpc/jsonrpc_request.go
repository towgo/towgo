package jsonrpc

import (
	"context"
	"encoding/json"
	"sync"
)

// Request is the runtime wrapper around a JSON-RPC request and response.
type Request struct {
	server   *JsonrpcServer
	request  *Jsonrpcrequest
	response *Jsonrpcresponse
	conn     JsonRpcConnection
	connView JsonRpcConnection
	ctx      context.Context

	// Route is the matched route for the current request.
	Route *Route

	handlers []HandlerFunc
	index    int
	result   any
	err      error
	aborted  bool

	values sync.Map
}

// newRequest creates the runtime request wrapper used by the server pipeline.
func newRequest(ctx context.Context, server *JsonrpcServer, rpcReq *Jsonrpcrequest, rpcResp *Jsonrpcresponse, conn JsonRpcConnection) *Request {
	return &Request{
		server:   server,
		request:  rpcReq,
		response: rpcResp,
		conn:     conn,
		ctx:      ctx,
	}
}

// Server returns the server currently executing the request.
func (r *Request) Server() *JsonrpcServer {
	return r.server
}

// GetRpcRequest returns the protocol request payload.
func (r *Request) GetRpcRequest() *Jsonrpcrequest {
	return r.request
}

// GetRpcResponse returns the protocol response payload being built.
func (r *Request) GetRpcResponse() *Jsonrpcresponse {
	return r.response
}

// Connection returns the underlying transport connection, if any.
func (r *Request) Connection() JsonRpcConnection {
	return r.conn
}

// AsConnection adapts the request to the legacy JsonRpcConnection interface.
func (r *Request) AsConnection() JsonRpcConnection {
	if r.connView == nil {
		r.connView = &requestConnection{request: r}
	}
	return r.connView
}

// Context returns the runtime context attached to the request.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// GetCtx returns the runtime context attached to the request.
func (r *Request) GetCtx() context.Context {
	return r.Context()
}

// SetCtx replaces the runtime context for following middleware and handlers.
func (r *Request) SetCtx(ctx context.Context) {
	r.ctx = ctx
}

// Method returns the JSON-RPC method name from the protocol request.
func (r *Request) Method() string {
	if r.request == nil {
		return ""
	}
	return r.request.Method
}

// Session returns the JSON-RPC session value from the protocol request.
func (r *Request) Session() string {
	if r.request == nil {
		return ""
	}
	return r.request.Session
}

// Params returns the raw JSON-RPC params value.
func (r *Request) Params() any {
	if r.request == nil {
		return nil
	}
	return r.request.Params
}

// ReadParams unmarshals the JSON-RPC params into each destination.
func (r *Request) ReadParams(destParams ...any) error {
	paramsBytes, err := json.Marshal(r.Params())
	if err != nil {
		return err
	}
	for _, dest := range destParams {
		if err := json.Unmarshal(paramsBytes, dest); err != nil {
			return err
		}
	}
	return nil
}

// SetResult stores a successful result without aborting the handler chain.
func (r *Request) SetResult(result any) {
	r.result = result
}

// GetResult returns the result stored by middleware or a handler.
func (r *Request) GetResult() any {
	return r.result
}

// WriteResult stores a result and aborts the remaining handler chain.
func (r *Request) WriteResult(result any) {
	r.SetResult(result)
	r.Abort()
}

// SetError writes a JSON-RPC error code and message to the response.
func (r *Request) SetError(code int64, msg string) {
	r.response.Error.Set(code, msg)
	r.err = &r.response.Error
}

// WithError stores an execution error for the response flush step.
func (r *Request) WithError(err error) {
	r.err = err
}

// GetError returns the execution error stored on the request.
func (r *Request) GetError() error {
	return r.err
}

// SetValue stores a request-scoped value.
func (r *Request) SetValue(key, value any) {
	r.values.Store(key, value)
}

// GetValue loads a request-scoped value.
func (r *Request) GetValue(key any) (value any, ok bool) {
	return r.values.Load(key)
}

// Abort stops the middleware and handler chain.
func (r *Request) Abort() {
	r.aborted = true
}

// Next executes the next middleware or handler in the chain.
func (r *Request) Next() {
	if r.aborted {
		return
	}
	if r.index >= len(r.handlers) {
		return
	}
	handler := r.handlers[r.index]
	r.index++
	handler(r)
}

// flush copies the stored result or error into the protocol response payload.
func (r *Request) flush() {
	if r.err != nil {
		if rpcErr, ok := r.err.(*Error); ok {
			r.response.Error = *rpcErr
		} else {
			r.response.Error.Set(500, r.err.Error())
		}
		return
	}
	if r.result != nil {
		r.response.Result = r.result
	}
}

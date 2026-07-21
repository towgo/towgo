package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
)

// Connection is an alias for JsonRpcConnection.
type Connection = JsonRpcConnection

// ConnectionHandlerFunc is an alias for Handler.
type ConnectionHandlerFunc = Handler

// ConnectionMiddlewareFunc is an alias for MiddlewareHandler.
type ConnectionMiddlewareFunc = MiddlewareHandler

// ExecConnection executes a request from a JsonRpcConnection and returns a response.
func (s *JsonrpcServer) ExecConnection(conn JsonRpcConnection) *Jsonrpcresponse {
	if conn == nil {
		resp := NewJsonrpcresponse()
		resp.Error.Set(500, "nil jsonrpc connection")
		return resp
	}
	return s.exec(conn.Context(), conn.GetRpcRequest(), conn.GetRpcResponse(), conn)
}

// ServeConnection executes a connection request and writes the response back.
func (s *JsonrpcServer) ServeConnection(conn JsonRpcConnection) {
	resp := s.ExecConnection(conn)
	if conn != nil {
		conn.WriteResponse(*resp)
	}
}

// MiddlewareConnection adapts connection-style middleware to request middleware.
func (s *JsonrpcServer) MiddlewareConnection(middlewares ...MiddlewareHandler) *JsonrpcServer {
	for _, middleware := range middlewares {
		m := middleware
		s.Middleware(func(r *Request) {
			m(r.AsConnection())
		})
	}
	return s
}

// UseConnection adapts connection-style middleware to request middleware.
func (s *JsonrpcServer) UseConnection(middlewares ...MiddlewareHandler) *JsonrpcServer {
	return s.MiddlewareConnection(middlewares...)
}

// HandleConnection registers a connection-style handler for one method.
func (s *JsonrpcServer) HandleConnection(method string, handler Handler, middlewares ...MiddlewareHandler) *Route {
	return s.Handle(method, wrapConnectionHandler(handler), wrapConnectionMiddlewares(middlewares...)...)
}

// MiddlewareConnection adapts connection-style middleware to this group.
func (g *RouterGroup) MiddlewareConnection(middlewares ...MiddlewareHandler) *RouterGroup {
	for _, middleware := range middlewares {
		m := middleware
		g.Middleware(func(r *Request) {
			m(r.AsConnection())
		})
	}
	return g
}

// UseConnection adapts connection-style middleware to this group.
func (g *RouterGroup) UseConnection(middlewares ...MiddlewareHandler) *RouterGroup {
	return g.MiddlewareConnection(middlewares...)
}

// HandleConnection registers a connection-style handler for one group method.
func (g *RouterGroup) HandleConnection(method string, handler Handler, middlewares ...MiddlewareHandler) *Route {
	return g.Handle(method, wrapConnectionHandler(handler), wrapConnectionMiddlewares(middlewares...)...)
}

// wrapConnectionHandler adapts a connection handler to the request pipeline.
func wrapConnectionHandler(handler Handler) HandlerFunc {
	return func(r *Request) {
		handler(r.AsConnection())
	}
}

// wrapConnectionMiddlewares adapts connection middleware to request middleware.
func wrapConnectionMiddlewares(middlewares ...MiddlewareHandler) []MiddlewareFunc {
	handlers := make([]MiddlewareFunc, 0, len(middlewares))
	for _, middleware := range middlewares {
		m := middleware
		handlers = append(handlers, func(r *Request) {
			m(r.AsConnection())
		})
	}
	return handlers
}

// requestConnection exposes one Request through the JsonRpcConnection API.
type requestConnection struct {
	request *Request
}

// GetRemoteAddr returns the remote peer address from the underlying connection.
func (c *requestConnection) GetRemoteAddr() string {
	if conn := c.request.Connection(); conn != nil {
		return conn.GetRemoteAddr()
	}
	return ""
}

// Read returns the JSON encoding of the protocol request.
func (c *requestConnection) Read() string {
	data, _ := json.Marshal(c.request.GetRpcRequest())
	return string(data)
}

// ReadParams unmarshals request params into each destination.
func (c *requestConnection) ReadParams(destParams ...interface{}) error {
	return c.request.ReadParams(destParams...)
}

// ReadResult unmarshals response result into each destination.
func (c *requestConnection) ReadResult(destResult ...interface{}) error {
	data, err := json.Marshal(c.request.GetRpcResponse().Result)
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

// SetResult stores a response result on the request.
func (c *requestConnection) SetResult(result interface{}) {
	c.request.SetResult(result)
}

// GetResult returns the response result stored on the request.
func (c *requestConnection) GetResult() interface{} {
	return c.request.GetResult()
}

// WriteResult stores a result and aborts the request pipeline.
func (c *requestConnection) WriteResult(result interface{}) {
	c.request.WriteResult(result)
}

// Write flushes the request response and aborts the pipeline.
func (c *requestConnection) Write() {
	c.request.flush()
	c.request.Abort()
}

// GetRpcRequest returns the protocol request payload.
func (c *requestConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.request.GetRpcRequest()
}

// GetRpcResponse returns the protocol response payload.
func (c *requestConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.request.GetRpcResponse()
}

// Push sends a one-way request through the underlying connection.
func (c *requestConnection) Push(request *Jsonrpcrequest) error {
	if conn := c.request.Connection(); conn != nil {
		return conn.Push(request)
	}
	return errors.New("jsonrpc connection is not available")
}

// Call delegates a JSON-RPC call to the underlying connection.
func (c *requestConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if conn := c.request.Connection(); conn != nil {
		conn.Call(rpcRequest, callback)
	}
}

// LinkType returns the underlying transport name.
func (c *requestConnection) LinkType() string {
	if conn := c.request.Connection(); conn != nil {
		return conn.LinkType()
	}
	return "local"
}

// IsConnected reports whether the underlying transport is connected.
func (c *requestConnection) IsConnected() bool {
	if conn := c.request.Connection(); conn != nil {
		return conn.IsConnected()
	}
	return true
}

// GUID returns the underlying connection identifier.
func (c *requestConnection) GUID() string {
	if conn := c.request.Connection(); conn != nil {
		return conn.GUID()
	}
	return ""
}

// EnableHealthCheck delegates health check startup to the connection.
func (c *requestConnection) EnableHealthCheck() {
	if conn := c.request.Connection(); conn != nil {
		conn.EnableHealthCheck()
	}
}

// DisableHealthCheck delegates health check shutdown to the connection.
func (c *requestConnection) DisableHealthCheck() {
	if conn := c.request.Connection(); conn != nil {
		conn.DisableHealthCheck()
	}
}

// WriteError stores an error and aborts the request pipeline.
func (c *requestConnection) WriteError(code int64, msg string) {
	c.request.SetError(code, msg)
	c.request.Abort()
}

// WriteResponse replaces the response payload and aborts the pipeline.
func (c *requestConnection) WriteResponse(resp Jsonrpcresponse) {
	*c.request.GetRpcResponse() = resp
	c.request.Abort()
}

// Close closes the underlying connection when one exists.
func (c *requestConnection) Close() {
	if conn := c.request.Connection(); conn != nil {
		conn.Close()
	}
}

// SetValue stores a request-scoped value.
func (c *requestConnection) SetValue(key string, value any) {
	c.request.SetValue(key, value)
}

// GetValue loads a request-scoped value.
func (c *requestConnection) GetValue(key string) (value any, ok bool) {
	return c.request.GetValue(key)
}

// Context returns the request runtime context.
func (c *requestConnection) Context() context.Context {
	return c.request.Context()
}

// WithContext replaces the request runtime context.
func (c *requestConnection) WithContext(ctx context.Context) {
	c.request.SetCtx(ctx)
}

// GetError returns the request execution error.
func (c *requestConnection) GetError() error {
	return c.request.GetError()
}

// WithError stores a request execution error.
func (c *requestConnection) WithError(err error) {
	c.request.WithError(err)
}

// Next continues the request pipeline.
func (c *requestConnection) Next() {
	c.request.Next()
}

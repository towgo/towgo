package jsonrpc

import "context"

// JsonRpcConnection abstracts the transport used by a JSON-RPC request.
type JsonRpcConnection interface {
	// GetRemoteAddr returns the remote peer address when the transport has one.
	GetRemoteAddr() string
	// Read returns the raw JSON-RPC message body.
	Read() string
	// ReadParams unmarshals request params into each destination.
	ReadParams(destParams ...interface{}) error
	// ReadResult unmarshals response result into each destination.
	ReadResult(destResult ...interface{}) error
	// SetResult stores a response result without writing immediately.
	SetResult(result interface{})
	// GetResult returns the stored response result.
	GetResult() interface{}
	// WriteResult stores a result and writes the response.
	WriteResult(result interface{})
	// Write writes the current response to the transport.
	Write()
	// GetRpcRequest returns the protocol request payload.
	GetRpcRequest() *Jsonrpcrequest
	// GetRpcResponse returns the protocol response payload.
	GetRpcResponse() *Jsonrpcresponse
	// Push sends a one-way JSON-RPC request when the transport supports it.
	Push(request *Jsonrpcrequest) error
	// Call sends a JSON-RPC request and invokes callback when a response arrives.
	Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection))
	// LinkType returns the transport name, such as local, http, tcp, or websocket.
	LinkType() string
	// IsConnected reports whether the transport is still connected.
	IsConnected() bool
	// GUID returns the connection identifier.
	GUID() string
	// EnableHealthCheck starts transport health checks when supported.
	EnableHealthCheck()
	// DisableHealthCheck stops transport health checks when supported.
	DisableHealthCheck()
	// WriteError writes a JSON-RPC error response.
	WriteError(code int64, msg string)
	// WriteResponse writes a complete JSON-RPC response.
	WriteResponse(resp Jsonrpcresponse)
	// Close closes the transport.
	Close()
	// SetValue stores a connection-scoped value.
	SetValue(key string, value any)
	// GetValue loads a connection-scoped value.
	GetValue(key string) (value any, ok bool)
	// Context returns the runtime context attached to the connection.
	Context() context.Context
	// WithContext replaces the runtime context attached to the connection.
	WithContext(ctx context.Context)
	// GetError returns the execution error stored on the connection.
	GetError() error
	// WithError stores an execution error on the connection.
	WithError(err error)
	// Next continues the middleware or handler chain.
	Next()
}

// Handler handles a request through the JsonRpcConnection compatibility API.
type Handler func(conn JsonRpcConnection)

// MiddlewareHandler wraps a connection handler and should call JsonRpcConnection.Next.
type MiddlewareHandler func(conn JsonRpcConnection)

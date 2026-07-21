package jsonrpc

import (
	"context"
	"sort"
	"sync"
	"time"
)

// HandlerFunc handles a normalized JSON-RPC request inside this server.
type HandlerFunc func(r *Request)

// MiddlewareFunc wraps a request handler and should call Request.Next to continue.
type MiddlewareFunc func(r *Request)

// RecoverHandler handles panics raised while a request chain is executing.
type RecoverHandler func(r *Request, v any)

// Req is a short alias for Request.
type Req = Request

// Middle is a short alias for MiddlewareFunc.
type Middle = MiddlewareFunc

// JsonrpcServer routes JSON-RPC requests to middleware and handlers.
type JsonrpcServer struct {
	mu             sync.RWMutex
	routes         map[string]*Route
	middlewares    []MiddlewareFunc
	recoverHandler RecoverHandler
}

// JsonRpcServer is an alternate spelling for JsonrpcServer.
type JsonRpcServer = JsonrpcServer

// Route stores one JSON-RPC method binding and its route-level middleware.
type Route struct {
	// Method is the normalized JSON-RPC method path.
	Method string
	// Handler is the terminal function for this route.
	Handler HandlerFunc
	// Middlewares are route-specific handlers that run before Handler.
	Middlewares []MiddlewareFunc
}

// NewJsonrpcServer creates a JSON-RPC server with an empty route table.
func NewJsonrpcServer() *JsonrpcServer {
	return &JsonrpcServer{
		routes:         make(map[string]*Route),
		recoverHandler: defaultRecoverHandler,
	}
}

// NewJsonRpcServer creates a JSON-RPC server using the alternate RPC spelling.
func NewJsonRpcServer() *JsonrpcServer {
	return NewJsonrpcServer()
}

// NewServer creates a JSON-RPC server.
func NewServer() *JsonrpcServer {
	return NewJsonrpcServer()
}

// Middleware appends global middleware to the server.
func (s *JsonrpcServer) Middleware(middlewares ...MiddlewareFunc) *JsonrpcServer {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middlewares = append(s.middlewares, middlewares...)
	return s
}

// Use appends global middleware to the server.
func (s *JsonrpcServer) Use(middlewares ...MiddlewareFunc) *JsonrpcServer {
	return s.Middleware(middlewares...)
}

// SetRecoverHandler replaces the panic recovery handler used by the server.
func (s *JsonrpcServer) SetRecoverHandler(handler RecoverHandler) *JsonrpcServer {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recoverHandler = handler
	return s
}

// Group creates a router group with a shared method prefix and middleware stack.
func (s *JsonrpcServer) Group(prefix string, groups ...func(group *RouterGroup)) *RouterGroup {
	group := &RouterGroup{
		server: s,
		prefix: cleanPrefix(prefix),
	}
	if len(groups) > 0 {
		for _, f := range groups {
			f(group)
		}
	}
	return group
}

// Bind registers functions or controller objects on the root router group.
func (s *JsonrpcServer) Bind(handlerOrObject ...any) *JsonrpcServer {
	s.Group("/").Bind(handlerOrObject...)
	return s
}

// BindObject registers functions or controller objects under a method prefix.
func (s *JsonrpcServer) BindObject(prefix string, handlerOrObject ...any) *JsonrpcServer {
	s.Group(prefix).Bind(handlerOrObject...)
	return s
}

// Handle registers a handler for one JSON-RPC method.
func (s *JsonrpcServer) Handle(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	return s.addRoute(cleanMethod(method), handler, middlewares...)
}

// ALL registers a handler for one JSON-RPC method.
func (s *JsonrpcServer) ALL(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	return s.Handle(method, handler, middlewares...)
}

// Handler registers a handler for one JSON-RPC method.
func (s *JsonrpcServer) Handler(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	return s.Handle(method, handler, middlewares...)
}

// Methods returns the registered JSON-RPC method names in sorted order.
func (s *JsonrpcServer) Methods() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	methods := make([]string, 0, len(s.routes))
	for method := range s.routes {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods
}

// Exec executes one JSON-RPC request with the provided runtime context.
func (s *JsonrpcServer) Exec(ctx context.Context, rpcReq *Jsonrpcrequest) *Jsonrpcresponse {
	return s.exec(ctx, rpcReq, nil, nil)
}

// exec runs the request chain using optional prebuilt response and connection state.
func (s *JsonrpcServer) exec(ctx context.Context, rpcReq *Jsonrpcrequest, rpcResp *Jsonrpcresponse, conn JsonRpcConnection) *Jsonrpcresponse {
	if ctx == nil {
		ctx = context.Background()
	}
	if rpcReq == nil {
		rpcReq = &Jsonrpcrequest{
			Jsonrpc: "2.0",
		}
	}
	if rpcResp == nil {
		rpcResp = NewJsonrpcresponse()
	}
	if rpcResp.Error.Code == 0 && rpcResp.Error.Message == "" {
		rpcResp.Error.Set(200, "")
	}
	rpcResp.Id = rpcReq.Id
	rpcResp.Timestampin = rpcReq.Timestampin
	rpcResp.Timestampout = currentTimestamp()

	req := newRequest(ctx, s, rpcReq, rpcResp, conn)
	defer func() {
		if v := recover(); v != nil {
			req.response.Error.Set(500, "recover exception")
			if s.recoverHandler != nil {
				s.recoverHandler(req, v)
			}
		}
	}()

	route := s.matchRoute(rpcReq.Method)
	if route == nil {
		rpcResp.Error.Set(-32601, "")
		return rpcResp
	}
	req.Route = route
	req.handlers = s.buildHandlers(route)
	req.Next()
	req.flush()
	return rpcResp
}

// addRoute stores a normalized method route.
func (s *JsonrpcServer) addRoute(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	route := &Route{
		Method:      method,
		Handler:     handler,
		Middlewares: append([]MiddlewareFunc(nil), middlewares...),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[method] = route
	return route
}

// matchRoute finds a route by normalized method name.
func (s *JsonrpcServer) matchRoute(method string) *Route {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.routes[cleanMethod(method)]
}

// buildHandlers assembles global, route, and terminal handlers.
func (s *JsonrpcServer) buildHandlers(route *Route) []HandlerFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()

	handlers := make([]HandlerFunc, 0, len(s.middlewares)+len(route.Middlewares)+1)
	for _, middleware := range s.middlewares {
		handlers = append(handlers, HandlerFunc(middleware))
	}
	for _, middleware := range route.Middlewares {
		handlers = append(handlers, HandlerFunc(middleware))
	}
	handlers = append(handlers, route.Handler)
	return handlers
}

// defaultRecoverHandler converts panics into a JSON-RPC server error.
func defaultRecoverHandler(r *Request, v any) {
	r.SetError(500, "recover exception")
}

// currentTimestamp returns the current wall-clock timestamp for responses.
func currentTimestamp() string {
	return time.Now().Format(time.RFC3339Nano)
}

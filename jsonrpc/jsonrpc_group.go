package jsonrpc

// RouterGroup groups JSON-RPC routes under a shared method prefix.
type RouterGroup struct {
	parent      *RouterGroup
	server      *JsonrpcServer
	prefix      string
	middlewares []MiddlewareFunc
}

// Group creates a child group that inherits the current group middleware.
func (g *RouterGroup) Group(prefix string, groups ...func(group *RouterGroup)) *RouterGroup {
	group := &RouterGroup{
		parent:      g,
		server:      g.server,
		prefix:      cleanPrefix(prefix),
		middlewares: append([]MiddlewareFunc(nil), g.middlewares...),
	}
	if len(groups) > 0 {
		for _, f := range groups {
			f(group)
		}
	}
	return group
}

// Middleware appends middleware to this group.
func (g *RouterGroup) Middleware(middlewares ...MiddlewareFunc) *RouterGroup {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

// Use appends middleware to this group.
func (g *RouterGroup) Use(middlewares ...MiddlewareFunc) *RouterGroup {
	return g.Middleware(middlewares...)
}

// Bind registers functions or controller objects on this group.
func (g *RouterGroup) Bind(handlerOrObject ...any) *RouterGroup {
	for _, item := range handlerOrObject {
		g.bind(item)
	}
	return g
}

// BindObject registers functions or controller objects under a child prefix.
func (g *RouterGroup) BindObject(prefix string, handlerOrObject ...any) *RouterGroup {
	g.Group(prefix).Bind(handlerOrObject...)
	return g
}

// Handle registers a handler on this group.
func (g *RouterGroup) Handle(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	allMiddlewares := make([]MiddlewareFunc, 0, len(g.middlewares)+len(middlewares))
	allMiddlewares = append(allMiddlewares, g.middlewares...)
	allMiddlewares = append(allMiddlewares, middlewares...)
	return g.server.addRoute(g.fullMethod(method), handler, allMiddlewares...)
}

// ALL registers a handler on this group.
func (g *RouterGroup) ALL(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	return g.Handle(method, handler, middlewares...)
}

// Handler registers a handler on this group.
func (g *RouterGroup) Handler(method string, handler HandlerFunc, middlewares ...MiddlewareFunc) *Route {
	return g.Handle(method, handler, middlewares...)
}

// fullPrefix returns the normalized prefix inherited from all parent groups.
func (g *RouterGroup) fullPrefix() string {
	if g == nil {
		return ""
	}
	var parts []string
	for cur := g; cur != nil; cur = cur.parent {
		if cur.prefix != "" {
			parts = append([]string{cur.prefix}, parts...)
		}
	}
	method := ""
	for _, part := range parts {
		method = joinMethod(method, part)
	}
	return method
}

// fullMethod joins this group prefix and a method path.
func (g *RouterGroup) fullMethod(method string) string {
	return joinMethod(g.fullPrefix(), method)
}

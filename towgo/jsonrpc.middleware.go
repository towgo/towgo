/*
json rpc 2.0 中间件系统
by:liangliangit
中间件采用洋葱模型执行

使用示例：

	towgo.Middleware(func(conn towgo.JsonRpcConnection) {
	    // 前置逻辑...
	    conn.Next()
	    // 后置逻辑：检查 handler 设置的错误
	    if err := conn.GetError(); err != nil {
	        // 自定义错误处理
	        conn.WriteError(500, err.Error())
	    }
	})
*/
package towgo

import (
	"github.com/gogf/gf/v2/errors/gerror"
)

type MiddlewareHandler func(conn JsonRpcConnection)
type RecoverHandler func(conn JsonRpcConnection, v any)
type Handler func(conn JsonRpcConnection)

var (
	middlewares    []MiddlewareHandler
	recoverHandler RecoverHandler          = DefaultRecoverHandler
	methods        map[string]*HandlerInfo = map[string]*HandlerInfo{}
)

// 默认 panic 处理器
func DefaultRecoverHandler(conn JsonRpcConnection, v any) {
	if err, ok := v.(error); ok && gerror.HasStack(err) {
		conn.WithError(gerror.Wrap(err, "recover"))
	} else {
		conn.WithError(gerror.Wrap(err, "recover exception"))
	}
}

// 注册中间件
func Middleware(middleware ...MiddlewareHandler) {
	middlewares = append(middlewares, middleware...)
}

// 设置 panic 处理器
func SetRecoverHandler(h RecoverHandler) {
	recoverHandler = h
}

// 注册处理函数
func Handle(method string, handler Handler) *HandlerInfo {
	info := &HandlerInfo{
		method:  method,
		handler: handler,
	}
	methods[method] = info
	return info
}

// 获取处理函数
func getHandler(method string) (Handler, bool) {
	if info, ok := methods[method]; ok {
		return info.handler, true
	}
	return nil, false
}

// 执行入口（内部使用）
func execHandler(conn JsonRpcConnection) {
	defer func() {
		if r := recover(); r != nil && recoverHandler != nil {
			recoverHandler(conn, r)
		}
	}()

	var index int
	var next func()
	next = func() {
		if conn == nil {
			return
		}
		if index < len(middlewares) {
			m := middlewares[index]
			index++
			m(conn)
		} else {
			// 执行 handler
			req := conn.GetRpcRequest()
			if req == nil || req.Method == "" {
				conn.WithError(gerror.New("method not found"))
				return
			}
			handler, ok := getHandler(req.Method)
			if !ok {
				// 尝试旧版 SetFunc 注册的 method
				if api, exists := funcs[req.Method]; exists {
					api.Exec(conn)
					return
				}
				conn.WithError(gerror.New("method not found"))
				return
			}
			handler(conn)
		}
	}

	// 设置 next 闭包
	switch c := conn.(type) {
	case *HttpRpcConnection:
		c.SetNextFunc(next)
	case *TcpRpcConnection:
		c.SetNextFunc(next)
	case *WebSocketRpcConnection:
		c.SetNextFunc(next)
	case *NilRpcConnection:
		c.SetNextFunc(next)
	case *LocalRpcConnection:
		c.SetNextFunc(next)
	}

	// 触发执行
	conn.Next()
}

// HandlerInfo 对象
type HandlerInfo struct {
	method  string
	handler Handler
}

func (h *HandlerInfo) Method() string {
	return h.method
}

func (h *HandlerInfo) Handler() Handler {
	return h.handler
}

// TestExec 执行请求（仅用于测试中间件）
func TestExec(conn JsonRpcConnection) {
	execHandler(conn)
}

// example/ws_transport_test.go
package main

import (
	"context"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport"
	"github.com/towgo/towgo/v2/transport/websocket"
)

func main() {
	// -------------------------- 启动WS服务端（纯传输） --------------------------

	// -------------------------- 服务端（用接口调用） --------------------------
	var server transport.ServerTransport = websocket.NewWSServerTransport()
	// 注册单请求处理器（接口方法）
	server.SetRequestHandler(func(req *protocol.Request) (*protocol.Response, bool, error) {
		fmt.Printf("服务端收到单请求：method=%s, params=%v, id=%v\n", req.Method, req.Params, req.ID)
		// 处理业务逻辑
		var result interface{}
		switch req.Method {
		case "add":
			params := req.Params.([]interface{})
			a := int(params[0].(float64))
			b := int(params[1].(float64))
			result = a + b
		case "sub":
			params := req.Params.([]interface{})
			a := int(params[0].(float64))
			b := int(params[1].(float64))
			result = a - b
		default:
			return &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeMethodNotFound, fmt.Sprintf("method %s not found", req.Method)),
				ID:      req.ID,
			}, true, nil
		}

		// 构造响应
		return &protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			Result:  result,
			ID:      req.ID,
		}, true, nil
	})

	// 启动服务（接口方法）

	server.Start(context.Background(), ":8080")

}

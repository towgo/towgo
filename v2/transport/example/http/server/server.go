package main

import (
	"context"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport/http"
)

func main() {
	server := http.NewHTTPServerTransport()
	// ✅ 正确：所有分支都返回protocol.Response
	server.SetRequestHandler(func(req *protocol.Request) (*protocol.Response, bool, error) {

		// 2. 处理业务逻辑
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
		case "log":
			// 通知请求：返回nil + false
			fmt.Printf("Log: %v\n", req.Params)
			return nil, false, nil
		default:
			return &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeMethodNotFound, "method not found"),
				ID:      req.ID,
			}, true, nil
		}

		// 3. 封装为规范的Response返回
		return &protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			Result:  result, // Result字段承载业务结果，而非直接返回裸值
			ID:      req.ID, // 必须带ID（非通知请求）
			Metadata: map[string]interface{}{
				"X-Request-ID": req.ID,
			},
		}, true, nil
	})

	ctx := context.Background()
	_ = server.Start(ctx, ":8080")

}

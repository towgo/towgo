package main

import (
	"context"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport/http"
)

func main() {

	// --------------- 启动HTTP客户端 ---------------
	// 1. 创建HTTP客户端传输实例

	// 2. 连接服务端
	ctx := context.Background()

	// 创建客户端
	client := http.NewHTTPClientTransport()
	client.Dial(ctx, "http://localhost:8080/jsonrpc")

	// 构造批量请求
	batchReq := protocol.BatchRequest{
		{JSONRPC: protocol.JSONRPCVersion,
			Method: "add",
			Params: []int{1, 2},
			ID:     "1",
		}, {JSONRPC: protocol.JSONRPCVersion,
			Method: "add",
			Params: []int{1, 6},
			ID:     "2",
		},
	}

	// 发送批量请求
	batchResp, err := client.SendBatchRequest(ctx, batchReq, true)
	if err != nil {
		// 错误处理
		panic(err)
	}

	// 处理响应
	for _, resp := range batchResp {
		// ...
		fmt.Printf("%+v \n", resp)
	}

}

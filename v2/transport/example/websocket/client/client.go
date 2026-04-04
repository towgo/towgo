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

	// -------------------------- 客户端（用接口调用） --------------------------
	var client transport.ClientTransport = websocket.NewWSClientTransport()
	// 连接服务端（接口方法）
	err := client.Dial(context.Background(), "ws://localhost:8080/jsonrpc/ws")
	if err != nil {
		fmt.Printf("连接失败：%v\n", err)
		return
	}
	defer client.Close(context.Background())

	// -------------------------- 测试1：单请求 --------------------------
	singleReq := &protocol.Request{
		JSONRPC: protocol.JSONRPCVersion,
		Method:  "add",
		Params:  []interface{}{10, 20},
		ID:      "req-1",
	}
	singleResp, err := client.SendRequest(context.Background(), singleReq, true)
	if err != nil {
		fmt.Printf("send single request failed: %v\n", err)
	} else {
		fmt.Printf("单请求响应：%+v\n", singleResp)
	}

	// -------------------------- 测试2：批量请求 --------------------------
	batchReq := protocol.BatchRequest{
		{
			JSONRPC: protocol.JSONRPCVersion,
			Method:  "add",
			Params:  []interface{}{100, 200},
			ID:      "req-2",
		},
		{
			JSONRPC: protocol.JSONRPCVersion,
			Method:  "sub",
			Params:  []interface{}{500, 100},
			ID:      "req-3",
		},
	}
	batchResp, err := client.SendBatchRequest(context.Background(), batchReq, true)
	if err != nil {
		fmt.Printf("send batch request failed: %v\n", err)
	} else {
		fmt.Printf("批量请求响应：%+v\n", batchResp)
	}
	for _, v := range batchResp {
		fmt.Printf("批量请求响应：%+v\n", v)
	}

}

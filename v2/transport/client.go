package transport

import (
	"context"
	"github.com/towgo/towgo/v2/protocol"
)

// ClientTransport 客户端传输层统一接口
// 所有传输协议的客户端必须实现此接口，核心能力：连接服务端、发送请求
type ClientTransport interface {
	// Dial 建立与服务端的连接（UDP无连接，此方法仅做初始化）
	// ctx: 上下文，控制连接超时
	// addr: 服务端地址（如"localhost:8080"、"ws://127.0.0.1:8081/ws"）
	// 返回值：连接失败返回error（如网络不可达）
	Dial(ctx context.Context, addr string) error

	// Close 关闭客户端连接（释放资源）
	// ctx: 上下文，控制关闭超时
	// 返回值：关闭失败返回error
	Close(ctx context.Context) error

	// Protocol 返回当前传输协议类型
	Protocol() Protocol

	// SendRequest 发送JSON-RPC请求（支持扩展后的带metadata请求）
	// ctx: 上下文，控制请求超时/取消
	// req: 扩展后的请求对象（带metadata）
	// waitResponse: 是否等待响应（通知请求设为false，无需返回响应）
	// 返回值：扩展后的响应对象（带metadata）、错误（如网络错误、解析错误）
	SendRequest(ctx context.Context, req *protocol.Request, waitResponse bool) (*protocol.Response, error)

	// SendBatchRequest 发送批量请求
	// ctx: 上下文（超时/取消）
	// batchReq: 批量请求
	// waitResponse: 是否等待响应（所有请求都是通知则设为false）
	// 返回值：批量响应、错误
	SendBatchRequest(ctx context.Context, batchReq protocol.BatchRequest, waitResponse bool) (protocol.BatchResponse, error)
}

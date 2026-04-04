package transport

import (
	"context"
	"github.com/towgo/towgo/v2/protocol"
)

// ServerTransport 服务端传输层统一接口
// 所有传输协议的服务端必须实现此接口，核心能力：启动/停止、处理请求
type ServerTransport interface {
	// Start 启动传输服务（监听指定地址）
	// ctx: 上下文，用于优雅关闭（如ctx.Done()触发停止）
	// addr: 监听地址（如":8080"、"0.0.0.0:9090"）
	// 返回值：启动失败返回error（如端口被占用）
	Start(ctx context.Context, addr string) error

	// Stop 停止传输服务（释放资源、关闭监听/连接）
	// ctx: 上下文，控制停止超时
	// 返回值：停止失败返回error（如资源释放失败）
	Stop(ctx context.Context) error

	// Protocol 返回当前传输协议类型（如ProtocolHTTP）
	Protocol() Protocol

	// SetRequestHandler 设置请求处理器（由上层protocol/server层实现）
	// 入参：扩展后的JSON-RPC请求对象（带metadata）
	// 返回值：扩展后的响应对象（带metadata）、是否需要返回响应（通知请求返回false）、错误
	SetRequestHandler(handler func(req *protocol.Request) (*protocol.Response, bool, error))

	// SetBatchRequestHandler ：批量请求处理器（可选实现，默认基于单请求处理器封装）
	// 入参：批量请求
	// 返回值：批量响应、是否需要返回响应（所有请求都是通知则返回false）、错误
	SetBatchRequestHandler(handler func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error))

	// GetDefaultBatchHandler ：获取默认批量处理器（基于单请求处理器自动封装，避免重复实现）
	GetDefaultBatchHandler() func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error)
}

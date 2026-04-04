// Package transport 定义JSON-RPC 2.0多传输协议的统一接口，
// 所有传输协议（HTTP/WS/TCP/UDP）均实现此接口，依赖protocol层（扩展后带metadata）
package transport

import "github.com/towgo/towgo/v2/protocol"

// Protocol 传输协议类型枚举（标识不同传输层）
type Protocol string

// 预定义支持的传输协议（后续可扩展如QUIC）
const (
	ProtocolHTTP Protocol = "http"      // HTTP/HTTPS
	ProtocolWS   Protocol = "websocket" // WebSocket/WSS
	ProtocolTCP  Protocol = "tcp"       // TCP
	ProtocolUDP  Protocol = "udp"       // UDP

	// ContentTypeJSON HTTP请求默认Content-Type（仅HTTP层使用）
	ContentTypeJSON      = "application/json"
	DefaultHandlerPath   = "/jsonrpc"
	DefaultWsHandlerPath = "/jsonrpc/ws"
)

// -------------------------- 新增：默认批量处理器（通用实现） --------------------------
// DefaultBatchHandler 基于单请求处理器生成批量处理器（所有传输层复用）
func DefaultBatchHandler(singleHandler func(req *protocol.Request) (*protocol.Response, bool, error)) func(protocol.BatchRequest) (protocol.BatchResponse, bool, error) {
	return func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error) {
		// 复用protocol层的HandleBatchRequest逻辑
		batchResp := protocol.HandleBatchRequest(batchReq, singleHandler)

		// 判断是否需要返回响应（批量响应非空则需要）
		needResponse := len(batchResp) > 0
		return batchResp, needResponse, nil
	}
}

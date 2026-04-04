// Package websocket WS客户端传输层（完整支持单/批量请求收发）
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSClientTransport WS客户端传输层（实现transport.ClientTransport接口）
type WSClientTransport struct {
	protocol transport.Protocol // 固定为websocket
	conn     *websocket.Conn    // WS连接
	addr     string             // 服务端地址（如ws://localhost:8080/ws）
	mu       sync.Mutex         // 保护conn的并发读写
}

// NewWSClientTransport 创建WS客户端传输实例
func NewWSClientTransport() *WSClientTransport {
	return &WSClientTransport{
		protocol: transport.ProtocolWS,
	}
}

// -------------------------- 实现transport.ClientTransport接口 --------------------------
// Protocol 返回传输协议类型
func (w *WSClientTransport) Protocol() transport.Protocol {
	return w.protocol
}

// Dial 建立连接（修复：设置超时）
func (w *WSClientTransport) Dial(ctx context.Context, addr string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("parse addr failed: %w", err)
	}

	// 修复：设置拨号超时
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial ws failed: %w", err)
	}

	// 设置连接超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// 心跳
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	w.conn = conn
	w.addr = addr
	return nil
}

// Close 关闭连接（接口要求）
func (w *WSClientTransport) Close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		// 发送关闭帧
		err := w.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
		if err != nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			fmt.Printf("write close message error: %v\n", err)
		}
		return w.conn.Close()
	}
	return nil
}

// SendRequest 发送单请求（修复：完善错误处理）
func (w *WSClientTransport) SendRequest(ctx context.Context, req *protocol.Request, waitResponse bool) (*protocol.Response, error) {
	// 1. 校验请求
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	if req.JSONRPC == "" {
		req.JSONRPC = protocol.JSONRPCVersion
	}

	// 2. 序列化请求为字节
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 3. 调用纯字节发送
	respBytes, err := w.SendBytes(ctx, reqBytes, waitResponse)
	if err != nil {
		return nil, fmt.Errorf("send bytes failed: %w", err)
	}
	if !waitResponse {
		return nil, nil
	}

	// 4. 反序列化为响应（修复：兼容单错误响应）
	var resp *protocol.Response
	var batchResp protocol.BatchResponse
	// 先尝试解析为单响应
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		// 再尝试解析为批量响应（兼容服务端的错误返回）
		if err2 := json.Unmarshal(respBytes, &batchResp); err2 == nil && len(batchResp) > 0 {
			return batchResp[0], nil
		}
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	return resp, nil
}

// SendBatchRequest 发送批量请求（修复：完善错误处理）
func (w *WSClientTransport) SendBatchRequest(ctx context.Context, batchReq protocol.BatchRequest, waitResponse bool) (protocol.BatchResponse, error) {
	// 1. 校验请求
	if len(batchReq) == 0 {
		return nil, fmt.Errorf("batch request is empty")
	}
	// 补全JSONRPC版本
	for _, req := range batchReq {
		if req != nil && req.JSONRPC == "" {
			req.JSONRPC = protocol.JSONRPCVersion
		}
	}

	// 2. 序列化批量请求为字节
	reqBytes, err := json.Marshal(batchReq)
	if err != nil {
		return nil, fmt.Errorf("marshal batch request failed: %w", err)
	}

	// 3. 纯字节发送
	respBytes, err := w.SendBytes(ctx, reqBytes, waitResponse)
	if err != nil {
		return nil, fmt.Errorf("send bytes failed: %w", err)
	}
	if !waitResponse {
		return nil, nil
	}

	// 4. 反序列化为批量响应（修复：兼容单错误响应）
	var batchResp protocol.BatchResponse
	var singleResp *protocol.Response
	// 先尝试解析为批量响应
	if err := json.Unmarshal(respBytes, &batchResp); err != nil {
		// 再尝试解析为单响应（服务端返回的错误）
		if err2 := json.Unmarshal(respBytes, &singleResp); err2 == nil {
			batchResp = protocol.BatchResponse{singleResp}
		} else {
			return nil, fmt.Errorf("unmarshal batch response failed: %w", err)
		}
	}

	return batchResp, nil
}

// SendBytes 纯传输层字节发送（核心，修复超时和错误处理）
func (w *WSClientTransport) SendBytes(ctx context.Context, reqBytes []byte, waitResponse bool) ([]byte, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn == nil {
		return nil, fmt.Errorf("ws connection not established")
	}

	// 刷新写超时
	w.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// 发送字节
	if err := w.conn.WriteMessage(websocket.TextMessage, reqBytes); err != nil {
		return nil, fmt.Errorf("write message failed: %w", err)
	}

	if !waitResponse {
		return nil, nil
	}

	// 刷新读超时
	w.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	// 读取响应字节
	mt, respBytes, err := w.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read message failed: %w", err)
	}

	if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
		return nil, fmt.Errorf("unsupported message type: %d", mt)
	}
	return respBytes, nil
}

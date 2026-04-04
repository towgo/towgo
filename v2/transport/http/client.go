// Package http 实现JSON-RPC 2.0的HTTP传输客户端，完全实现transport.ClientTransport接口
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport"
	"net/http"
)

// HTTPClientTransport HTTP客户端传输实现（满足transport.ClientTransport接口）
type HTTPClientTransport struct {
	protocol transport.Protocol // 协议类型，固定为HTTP
	client   *http.Client       // 底层HTTP客户端实例
	addr     string             // 服务端地址（如"http://localhost:8080"）
}

// NewHTTPClientTransport 创建HTTP客户端传输实例（对外暴露的构造函数）
func NewHTTPClientTransport() *HTTPClientTransport {
	return &HTTPClientTransport{
		protocol: transport.ProtocolHTTP,
		client:   &http.Client{ // 可后续通过配置自定义超时、代理等
			// Timeout: time.Second * 10, // 示例：设置10秒超时
		},
	}
}

// -------------------------- 实现transport.ClientTransport接口 --------------------------
// Protocol 返回传输协议类型
func (h *HTTPClientTransport) Protocol() transport.Protocol {
	return h.protocol
}

// Dial 初始化客户端（HTTP无长连接，仅记录服务端地址）
func (h *HTTPClientTransport) Dial(ctx context.Context, addr string) error {
	h.addr = addr // 保存服务端地址（如":8080"会自动补全为"http://:8080"）
	return nil
}

// Close 关闭客户端（HTTP无长连接，空实现）
func (h *HTTPClientTransport) Close(ctx context.Context) error {
	return nil
}

// SendRequest 发送JSON-RPC请求（带metadata），并可选等待响应
func (h *HTTPClientTransport) SendRequest(ctx context.Context, req *protocol.Request, waitResponse bool) (*protocol.Response, error) {
	// 1. 校验服务端地址是否已设置
	if h.addr == "" {
		return nil, fmt.Errorf("client not dialed (server address is empty)") // 补充import "fmt"
	}

	// 2. 序列化请求为JSON（包含metadata）
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// 3. 构造HTTP POST请求
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		h.addr,
		bytes.NewBuffer(reqBytes), // 请求体为JSON
	)
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	// 设置Content-Type（必须）
	httpReq.Header.Set("Content-Type", transport.ContentTypeJSON)

	// 4. 发送HTTP请求
	httpResp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send http request failed: %w", err)
	}
	defer httpResp.Body.Close() // 关闭响应体

	// 5. 通知请求（无需等待响应）直接返回nil
	if !waitResponse {
		return nil, nil
	}

	// 6. 解析响应为扩展后的JSON-RPC Response（带metadata）
	var resp protocol.Response
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	// 7. 返回响应
	return &resp, nil
}

// transport/http/client.go - 重写SendBatchRequest
func (h *HTTPClientTransport) SendBatchRequest(ctx context.Context, batchReq protocol.BatchRequest, waitResponse bool) (protocol.BatchResponse, error) {
	// 1. 原有校验地址/序列化逻辑不变
	if h.addr == "" {
		return nil, fmt.Errorf("client not dialed (server address is empty)")
	}
	reqBytes, err := json.Marshal(batchReq)
	if err != nil {
		return nil, fmt.Errorf("marshal batch request failed: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, h.addr, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", transport.ContentTypeJSON)

	// 2. 发送请求（原有逻辑）
	httpResp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send batch request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// 3. 无需响应则返回nil（原有逻辑）
	if !waitResponse {
		return nil, nil
	}

	// -------------------------- 核心修复：先解析为interface{} 判断类型 --------------------------
	var rawResp interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("unmarshal batch response failed: %w", err)
	}

	// 4. 分支1：响应是数组（正常批量响应）
	if respArray, ok := rawResp.([]interface{}); ok {
		var batchResp protocol.BatchResponse
		// 手动解析每个元素到Response
		for _, item := range respArray {
			itemBytes, _ := json.Marshal(item)
			var resp protocol.Response
			if err := json.Unmarshal(itemBytes, &resp); err != nil {
				return nil, fmt.Errorf("unmarshal batch item failed: %w", err)
			}
			batchResp = append(batchResp, &resp)
		}
		return batchResp, nil
	}

	// 5. 分支2：响应是对象（服务端返回单个错误/单个响应）
	if respObj, ok := rawResp.(map[string]interface{}); ok {
		// 解析为单个Response，包装成BatchResponse返回
		itemBytes, _ := json.Marshal(respObj)
		var singleResp protocol.Response
		if err := json.Unmarshal(itemBytes, &singleResp); err != nil {
			return nil, fmt.Errorf("unmarshal single response failed: %w", err)
		}
		return protocol.BatchResponse{&singleResp}, nil
	}

	// 6. 分支3：响应是数字/空值（非法响应，返回错误）
	return nil, fmt.Errorf("invalid batch response type: %+v(%T) (expected array/object)", rawResp, rawResp)
}

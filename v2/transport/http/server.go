// Package http 实现JSON-RPC 2.0的HTTP传输服务端，完全实现transport.ServerTransport接口
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport"
)

// HTTPServerTransport HTTP服务端传输实现（满足transport.ServerTransport接口）
type HTTPServerTransport struct {
	protocol     transport.Protocol                                        // 协议类型，固定为HTTP
	server       *http.Server                                              // 底层HTTP服务实例
	handler      func(*protocol.Request) (*protocol.Response, bool, error) // 请求处理器
	batchHandler func(protocol.BatchRequest) (protocol.BatchResponse, bool, error)
	httpMux      *http.ServeMux
	handlerPath  string
	mu           sync.RWMutex // 并发安全锁（多协程读写handler）
}

func (h *HTTPServerTransport) SetHandlerPath(path string) {
	h.handlerPath = path
}

// NewHTTPServerTransport 创建HTTP服务端传输实例（对外暴露的构造函数）
func NewHTTPServerTransport() *HTTPServerTransport {
	return &HTTPServerTransport{
		protocol:    transport.ProtocolHTTP, // 固定为HTTP协议
		handlerPath: transport.DefaultHandlerPath,
		httpMux:     http.NewServeMux(),
	}
}
func (h *HTTPServerTransport) BindHttpMux(mux *http.ServeMux) {
	mux.HandleFunc(h.handlerPath, h.handleRPCRequest)
}
func (h *HTTPServerTransport) GetHttpMux() *http.ServeMux {
	return h.httpMux
}

// Protocol 返回传输协议类型
func (h *HTTPServerTransport) Protocol() transport.Protocol {
	return h.protocol
}

// SetRequestHandler 设置请求处理器（上层业务逻辑入口）
func (h *HTTPServerTransport) SetRequestHandler(handler func(*protocol.Request) (*protocol.Response, bool, error)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handler = handler // 保存处理器，并发安全
}

// Start 启动HTTP服务（监听地址，处理JSON-RPC请求）
func (h *HTTPServerTransport) Start(ctx context.Context, addr string) error {
	// 1. 注册JSON-RPC请求处理函数

	h.BindHttpMux(h.GetHttpMux())
	// 2. 初始化HTTP服务
	h.server = &http.Server{
		Addr:    addr,
		Handler: h.GetHttpMux(),
	}

	// 3. 监听上下文取消信号，实现优雅关闭
	go func() {
		<-ctx.Done() // 当ctx被取消（如调用Stop），触发Shutdown
		_ = h.server.Shutdown(context.Background())
	}()

	// 4. 启动HTTP服务（阻塞直到停止）
	return h.server.ListenAndServe()
}

// Stop 停止HTTP服务（优雅关闭，不中断正在处理的请求）
func (h *HTTPServerTransport) Stop(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.server != nil {
		return h.server.Shutdown(ctx) // 优雅关闭
	}
	return nil
}

// transport/http/server.go - 重写handleRPCRequest函数
func (h *HTTPServerTransport) handleRPCRequest(w http.ResponseWriter, r *http.Request) {
	// 1. 基础校验（方法/Content-Type）→ 原有逻辑不变
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte("JSON-RPC over HTTP only supports POST method"))
		return
	}
	if r.Header.Get("Content-Type") != transport.ContentTypeJSON {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		_, _ = w.Write([]byte("Content-Type must be application/json"))
		return
	}

	// 2. 读取请求体并解析为interface{}（识别单/批量请求）
	var rawData interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		// 解析失败 → 返回单个解析错误响应
		resp := &protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			Error:   protocol.NewStandardError(protocol.CodeParseError, err.Error()),
		}
		w.Header().Set("Content-Type", transport.ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	defer r.Body.Close()

	// 3. 获取请求处理器（并发安全）
	h.mu.RLock()
	handler := h.handler
	h.mu.RUnlock()
	if handler == nil {
		resp := &protocol.Response{
			JSONRPC: protocol.JSONRPCVersion,
			Error:   protocol.NewStandardError(protocol.CodeInternalError, "request handler not set"),
		}
		w.Header().Set("Content-Type", transport.ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// 4. 分支1：单个请求（JSON对象）
	if rawObj, ok := rawData.(map[string]interface{}); ok {
		// 解析为单请求（适配string ID）
		req, err := parseSingleRequest(rawObj)
		if err != nil {
			resp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInvalidRequest, err.Error()),
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// 校验+处理单请求（原有逻辑）
		valid, reqErr := protocol.ValidateRequest(req)
		if !valid {
			resp := &protocol.Response{JSONRPC: protocol.JSONRPCVersion, Error: reqErr, ID: req.ID}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		resp, needResponse, err := handler(req)
		if err != nil {
			resp = &protocol.Response{JSONRPC: protocol.JSONRPCVersion, Error: protocol.NewStandardError(protocol.CodeInternalError, err.Error()), ID: req.ID}
		}

		if !needResponse {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", transport.ContentTypeJSON)
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// 5. 分支2：批量请求（JSON数组）
	if rawArray, ok := rawData.([]interface{}); ok {
		// 解析为批量请求（适配string ID）
		batchReq, err := parseBatchRequest(rawArray)
		if err != nil {
			resp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInvalidRequest, err.Error()),
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// 处理批量请求
		batchResp := protocol.HandleBatchRequest(batchReq, handler)

		// 规则1：空数组 → 返回单个无效请求错误
		if len(rawArray) == 0 {
			resp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInvalidRequest, "batch request cannot be empty"),
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// 规则4：所有请求都非法 → 返回单个错误响应
		if batchResp == nil {
			resp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInvalidRequest, "all requests in batch are invalid"),
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// 规则3：返回批量响应数组（仅含带ID请求的响应）
		if len(batchResp) > 0 {
			w.Header().Set("Content-Type", transport.ContentTypeJSON)
			_ = json.NewEncoder(w).Encode(batchResp)
			return
		}

		// 所有请求都是通知请求 → 返回204 No Content
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 6. 既不是对象也不是数组 → 无效请求
	resp := &protocol.Response{
		JSONRPC: protocol.JSONRPCVersion,
		Error:   protocol.NewStandardError(protocol.CodeInvalidRequest, "request must be JSON object (single) or array (batch)"),
	}
	w.Header().Set("Content-Type", transport.ContentTypeJSON)
	_ = json.NewEncoder(w).Encode(resp)
}

// 实现SetBatchRequestHandler：允许自定义批量处理器（覆盖默认）
func (h *HTTPServerTransport) SetBatchRequestHandler(handler func(protocol.BatchRequest) (protocol.BatchResponse, bool, error)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.batchHandler = handler
}

// 实现GetDefaultBatchHandler：返回基于单请求的默认批量处理器
func (h *HTTPServerTransport) GetDefaultBatchHandler() func(protocol.BatchRequest) (protocol.BatchResponse, bool, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return transport.DefaultBatchHandler(h.handler)
}

// -------------------------- 辅助函数：解析单请求（适配string ID） --------------------------
func parseSingleRequest(rawObj map[string]interface{}) (*protocol.Request, error) {
	req := &protocol.Request{
		JSONRPC: rawObj["jsonrpc"].(string),
		Method:  rawObj["method"].(string),
		Params:  rawObj["params"],
		ID:      rawObj["id"].(string),
	}

	// 解析Metadata
	if md, ok := rawObj["metadata"].(map[string]interface{}); ok {
		req.Metadata = md
	}

	return req, nil
}

// -------------------------- 辅助函数：解析批量请求（适配string ID） --------------------------
func parseBatchRequest(rawArray []interface{}) (protocol.BatchRequest, error) {
	batchReq := make(protocol.BatchRequest, 0, len(rawArray))
	for _, item := range rawArray {
		if item == nil {
			batchReq = append(batchReq, nil)
			continue
		}
		rawObj, ok := item.(map[string]interface{})
		if !ok {
			batchReq = append(batchReq, nil)
			continue
		}
		req, err := parseSingleRequest(rawObj)
		if err != nil {
			batchReq = append(batchReq, nil)
			continue
		}
		batchReq = append(batchReq, req)
	}
	return batchReq, nil
}

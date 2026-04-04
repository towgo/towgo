// Package websocket 实现JSON-RPC 2.0的WebSocket传输层（完整支持单/批量请求）
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/v2/protocol"
	"github.com/towgo/towgo/v2/transport"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 定义WS升级器（可配置参数）
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有跨域，生产环境需配置白名单
	},
}

// DataHandler WS传输层的纯数据回调：入参是请求字节，出参是响应字节
type DataHandler func(reqBytes []byte) (respBytes []byte, needResponse bool, err error)

// WSServerTransport WS服务端传输层（实现transport.ServerTransport接口）
type WSServerTransport struct {
	protocol      transport.Protocol // 固定为websocket
	listener      *http.Server       // 底层HTTP服务（用于WS升级）
	dataHandler   DataHandler        // 纯数据回调（由应用层注册）
	handlerPath   string
	connManager   sync.Map // 管理所有WS连接（conn -> struct{}{}）
	singleHandler func(req *protocol.Request) (*protocol.Response, bool, error)
	batchHandler  func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error)
	mu            sync.RWMutex // 保护dataHandler/处理器
}

// NewWSServerTransport 创建WS服务端传输实例
func NewWSServerTransport() *WSServerTransport {
	return &WSServerTransport{
		protocol:    transport.ProtocolWS,
		handlerPath: transport.DefaultWsHandlerPath,
	}
}

func (w *WSServerTransport) SetHandlerPath(path string) {
	w.handlerPath = path
}

// -------------------------- 实现transport.ServerTransport接口 --------------------------
// Protocol 返回传输协议类型
func (w *WSServerTransport) Protocol() transport.Protocol {
	return w.protocol
}

func (w *WSServerTransport) BindHttpMux(mux *http.ServeMux) {
	mux.HandleFunc(w.handlerPath, w.handleWSUpgrade)
}

// Start 启动WS服务（监听地址，处理WS升级）
func (w *WSServerTransport) Start(ctx context.Context, addr string) error {
	// 注册WS升级路由
	mux := http.NewServeMux()
	w.BindHttpMux(mux)
	// 初始化HTTP服务
	w.listener = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 优雅关闭
	go func() {
		<-ctx.Done()
		_ = w.listener.Shutdown(ctx)
		// 关闭所有WS连接
		w.connManager.Range(func(key, value interface{}) bool {
			conn := key.(*websocket.Conn)
			_ = conn.Close()
			w.connManager.Delete(conn)
			return true
		})
	}()

	// 启动服务（阻塞）
	return w.listener.ListenAndServe()
}

// Stop 停止WS服务
func (w *WSServerTransport) Stop(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.listener != nil {
		return w.listener.Shutdown(ctx)
	}
	return nil
}

// SetRequestHandler 设置单请求处理器（接口要求）
// 修复点：保证dataHandler能正确处理单请求，且不被批量处理器覆盖
func (w *WSServerTransport) SetRequestHandler(handler func(req *protocol.Request) (*protocol.Response, bool, error)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.singleHandler = handler

	// 初始化通用dataHandler（兼容单/批量请求）
	w.initDataHandler()
}

// SetBatchRequestHandler 设置批量请求处理器（接口要求）
// 修复点：避免递归调用，兜底默认批量处理器
func (w *WSServerTransport) SetBatchRequestHandler(handler func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.batchHandler = handler

	// 初始化通用dataHandler（兼容单/批量请求）
	w.initDataHandler()
}

// initDataHandler 初始化通用dataHandler（核心修复：自动识别单/批量请求）
func (w *WSServerTransport) initDataHandler() {
	if w.dataHandler != nil {
		return // 已初始化，避免覆盖
	}

	w.dataHandler = func(reqBytes []byte) (respBytes []byte, needResponse bool, err error) {
		// 第一步：尝试解析为批量请求
		var batchReq protocol.BatchRequest
		if err := json.Unmarshal(reqBytes, &batchReq); err == nil && len(batchReq) > 0 {
			// 批量请求逻辑
			batchHandler := w.batchHandler
			if batchHandler == nil {
				// 兜底：使用默认批量处理器
				batchHandler = transport.DefaultBatchHandler(w.singleHandler)
			}
			batchResp, need, err := batchHandler(batchReq)
			if err != nil {
				// 返回JSON-RPC标准错误
				errorResp := &protocol.Response{
					JSONRPC: protocol.JSONRPCVersion,
					Error:   protocol.NewStandardError(protocol.CodeInternalError, err.Error()),
				}
				respBytes, _ = json.Marshal(errorResp)
				return respBytes, true, err
			}
			if !need {
				return nil, false, nil
			}
			respBytes, err = json.Marshal(batchResp)
			return respBytes, need, err
		}

		// 第二步：解析为单请求
		var req *protocol.Request
		if err := json.Unmarshal(reqBytes, &req); err != nil {
			// 修复点：返回JSON-RPC解析错误（而非空）
			errorResp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeParseError, fmt.Sprintf("parse request failed: %v", err)),
			}
			respBytes, _ = json.Marshal(errorResp)
			return respBytes, true, err
		}

		// 单请求逻辑
		if w.singleHandler == nil {
			errorResp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInternalError, "single request handler not set"),
				ID:      req.ID,
			}
			respBytes, _ = json.Marshal(errorResp)
			return respBytes, true, fmt.Errorf("single handler not set")
		}

		resp, need, err := w.singleHandler(req)
		if err != nil {
			errorResp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInternalError, err.Error()),
				ID:      req.ID,
			}
			respBytes, _ = json.Marshal(errorResp)
			return respBytes, true, err
		}
		if !need {
			return nil, false, nil
		}
		respBytes, err = json.Marshal(resp)
		return respBytes, need, err
	}
}

// GetDefaultBatchHandler 获取默认批量处理器（接口要求）
func (w *WSServerTransport) GetDefaultBatchHandler() func(batchReq protocol.BatchRequest) (protocol.BatchResponse, bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return transport.DefaultBatchHandler(w.singleHandler)
}

// SetDataHandler 注册纯数据回调（兼容手动注册）
func (w *WSServerTransport) SetDataHandler(handler DataHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dataHandler = handler
}

// handleWSUpgrade 处理HTTP升级为WS连接（核心修复：超时+规范响应）
func (w *WSServerTransport) handleWSUpgrade(writer http.ResponseWriter, req *http.Request) {
	// 升级HTTP连接为WS连接
	conn, err := upgrader.Upgrade(writer, req, nil)
	if err != nil {
		return
	}
	// 修复点：设置连接超时（避免卡死）
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	// 心跳处理（可选，防止连接被断开）
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	w.connManager.Store(conn, struct{}{})
	defer func() {
		_ = conn.Close()
		w.connManager.Delete(conn)
	}()

	// 循环读取WS消息
	for {
		mt, msgBytes, err := conn.ReadMessage()
		if err != nil {
			// 正常断开不报错
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				break
			}
			// 其他错误打印日志
			fmt.Printf("read message error: %v\n", err)
			break
		}

		// 仅处理文本/二进制帧
		if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
			continue
		}

		// 获取dataHandler
		w.mu.RLock()
		handler := w.dataHandler
		w.mu.RUnlock()
		if handler == nil {
			// 修复点：返回规范的错误响应（而非空字节）
			errorResp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInternalError, "data handler not set"),
			}
			errorBytes, _ := json.Marshal(errorResp)
			_ = conn.WriteMessage(websocket.TextMessage, errorBytes)
			continue
		}

		// 调用回调处理请求
		respBytes, needResponse, err := handler(msgBytes)
		if err != nil {
			// 错误响应
			errorResp := &protocol.Response{
				JSONRPC: protocol.JSONRPCVersion,
				Error:   protocol.NewStandardError(protocol.CodeInternalError, fmt.Sprintf("handler error: %v", err)),
			}
			errorBytes, _ := json.Marshal(errorResp)
			_ = conn.WriteMessage(websocket.TextMessage, errorBytes)
			continue
		}

		if needResponse && len(respBytes) > 0 {
			// 发送响应（修复：检查发送错误）
			if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
				fmt.Printf("write message error: %v\n", err)
				break
			}
		}
	}
}

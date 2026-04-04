package protocol

import "fmt"

// protocol/response.go

// Response 扩展后的JSON-RPC 2.0响应对象
type Response struct {
	JSONRPC  string                 `json:"jsonrpc"`            // 必选
	Result   interface{}            `json:"result,omitempty"`   // 二选一
	Error    Error                  `json:"error,omitempty"`    // 二选一
	ID       string                 `json:"id"`                 // 必选
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 新增：响应元数据（类比HTTP响应头）
}

// Error（不变）
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BatchResponse（不变）
type BatchResponse []*Response

// NewError / NewStandardError（不变）

// NewError 创建自定义JSON-RPC错误对象
func NewError(code int, message string, data interface{}) Error {
	return Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// NewStandardError 创建协议标准错误（基于预定义的错误码）
func NewStandardError(code int, data interface{}) Error {
	var msg string
	switch code {
	case CodeParseError:
		msg = MsgParseError
	case CodeInvalidRequest:
		msg = MsgInvalidRequest
	case CodeMethodNotFound:
		msg = MsgMethodNotFound
	case CodeInvalidParams:
		msg = MsgInvalidParams
	case CodeInternalError:
		msg = MsgInternalError
	default:
		msg = fmt.Sprintf("unknown standard error code: %d", code)
	}
	return NewError(code, msg, data)
}

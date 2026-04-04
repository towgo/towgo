package protocol

// protocol/request.go

// Request 扩展后的JSON-RPC 2.0请求对象
// 新增Metadata字段（可选，JSON Object），完全向后兼容
type Request struct {
	JSONRPC  string                 `json:"jsonrpc"`            // 必选，固定为2.0
	Method   string                 `json:"method"`             // 必选
	Params   interface{}            `json:"params,omitempty"`   // 可选
	ID       string                 `json:"id,omitempty"`       // 可选（通知请求无）
	Metadata map[string]interface{} `json:"metadata,omitempty"` // 新增：元数据（类比HTTP Header）
}

// IsNotification（不变）
func IsNotification(req *Request) bool {
	return req.ID == ""
}

// BatchRequest（不变）
type BatchRequest []*Request

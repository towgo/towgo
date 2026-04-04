// Package protocol 实现JSON-RPC 2.0核心协议规范，与传输层无关
package protocol

// JSONRPCVersion JSON-RPC 2.0 固定版本标识
const JSONRPCVersion = "2.0"

// 协议标准错误码（JSON-RPC 2.0 保留，负值为协议级错误）
const (
	CodeParseError     = -32700 // 解析错误：JSON格式无效
	CodeInvalidRequest = -32600 // 无效请求：JSON格式合法但不符合协议规范
	CodeMethodNotFound = -32601 // 方法不存在：服务端未注册该方法
	CodeInvalidParams  = -32602 // 参数错误：参数格式/类型不匹配
	CodeInternalError  = -32603 // 内部错误：服务端方法执行异常
)

// 标准错误信息（对应错误码的默认描述）
const (
	MsgParseError     = "Parse error"
	MsgInvalidRequest = "Invalid Request"
	MsgMethodNotFound = "Method not found"
	MsgInvalidParams  = "Invalid params"
	MsgInternalError  = "Internal error"
)

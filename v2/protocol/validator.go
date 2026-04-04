package protocol

import "fmt"

// ValidateRequest 校验单个请求是否符合JSON-RPC 2.0基础规范
// 返回值：是否合法、非法时返回标准错误对象
func ValidateRequest(req *Request) (bool, Error) {

	// 1. 校验版本号
	if req.JSONRPC != JSONRPCVersion {
		return false, NewStandardError(CodeInvalidRequest,
			fmt.Sprintf("invalid jsonrpc version: expected '%s', got '%s'", JSONRPCVersion, req.JSONRPC))
	}
	// 2. 校验方法名（非空 + 禁止以rpc.开头）
	if req.Method == "" {
		return false, NewStandardError(CodeInvalidRequest, "method name is empty")
	}
	if len(req.Method) >= 4 && req.Method[:4] == "rpc." {
		return false, NewStandardError(CodeInvalidRequest, "method name cannot start with 'rpc.' (reserved by protocol)")
	}

	return true, Error{}
}

// ValidateBatchRequest 校验批量请求是否符合规范
// 批量请求不能是空数组，且每个子请求需满足基础规范
func ValidateBatchRequest(batchReq BatchRequest) (bool, Error) {
	// 1. 批量请求不能是空数组
	if len(batchReq) == 0 {
		return false, NewStandardError(CodeInvalidRequest, "batch request cannot be empty")
	}

	// 2. 逐个校验子请求（只要有一个请求格式合法，批量响应就需要返回；全非法则返回单个错误响应）
	hasValidRequest := false
	for _, req := range batchReq {
		if req == nil {
			continue
		}
		valid, _ := ValidateRequest(req)
		if valid {
			hasValidRequest = true
			break
		}
	}

	// 3. 全非法的批量请求，返回单个错误响应
	if !hasValidRequest {
		return false, NewStandardError(CodeInvalidRequest, "all requests in batch are invalid")
	}

	return true, Error{}
}

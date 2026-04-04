package protocol

import "fmt"

// HandleBatchRequest 处理批量请求（核心逻辑，适配string ID）
// singleHandler：处理单个合法请求的函数（和单请求处理器一致）
func HandleBatchRequest(batchReq BatchRequest, singleHandler func(*Request) (*Response, bool, error)) BatchResponse {
	var batchResp BatchResponse

	// 规则1：空数组 → 返回单个无效请求错误（后续在传输层处理）
	if len(batchReq) == 0 {
		return nil
	}

	// 规则2：逐个处理子请求，保持顺序
	hasValidRequest := false // 标记是否有合法请求
	for _, req := range batchReq {
		if req == nil {
			// 空请求 → 无效请求错误
			resp := &Response{
				JSONRPC: JSONRPCVersion,
				Error:   NewStandardError(CodeInvalidRequest, "null request in batch"),
			}
			batchResp = append(batchResp, resp)
			continue
		}

		// 校验单个请求（包括string ID的校验）
		valid, reqErr := ValidateRequest(req)
		fmt.Println("校验单个请求valid", valid, reqErr)

		if !valid {
			// 非法请求 → 返回错误响应
			resp := &Response{
				JSONRPC: JSONRPCVersion,
				Error:   reqErr,
				ID:      req.ID, // 即使非法，也返回原ID（规范要求）
			}
			batchResp = append(batchResp, resp)
			continue
		}

		hasValidRequest = true
		// 规则3：通知请求（ID=null）→ 执行但不加入响应
		if IsNotification(req) {
			singleHandler(req)
			continue
		}

		// 合法带ID请求 → 执行并加入响应
		resp, _, err := singleHandler(req)
		if err != nil {
			resp = &Response{
				JSONRPC: JSONRPCVersion,
				Error:   NewStandardError(CodeInternalError, err.Error()),
				ID:      req.ID,
			}
		}
		batchResp = append(batchResp, resp)
	}

	// 规则4：所有请求都非法 → 返回nil（传输层处理为单个错误响应）
	if !hasValidRequest {
		return nil
	}

	return batchResp
}

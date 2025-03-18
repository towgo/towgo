package accessGateway

/*
统一接入网关 管理API
by:liangliangit
*/
import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
)

func init() {
	jsonrpc.SetFunc("/unauthorizedMethod/create", unauthorizedMethod_create)
	jsonrpc.SetFunc("/unauthorizedMethod/delete", unauthorizedMethod_delete)
	jsonrpc.SetFunc("/unauthorizedMethod/update", unauthorizedMethod_update)
	jsonrpc.SetFunc("/unauthorizedMethod/detail", unauthorizedMethod_detail)
	jsonrpc.SetFunc("/unauthorizedMethod/list", unauthorizedMethod_list)
}

func unauthorizedMethod_create(rpcConn jsonrpc.JsonRpcConnection) {
	model := UnauthorizedMethod{}
	rpcConn.ReadParams(&model)
	_, err := basedboperat.Create(&model)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func unauthorizedMethod_update(rpcConn jsonrpc.JsonRpcConnection) {
	model := UnauthorizedMethod{}
	selectMap := map[string]interface{}{}
	rpcConn.ReadParams(&model, &selectMap)
	err := basedboperat.Update(&model, selectMap, "id = ?", model.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func unauthorizedMethod_delete(rpcConn jsonrpc.JsonRpcConnection) {
	var model UnauthorizedMethod
	rpcConn.ReadParams(&model)

	count, err := basedboperat.Delete(&model, model.ID, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(count)
}

func unauthorizedMethod_detail(rpcConn jsonrpc.JsonRpcConnection) {
	var model UnauthorizedMethod
	rpcConn.ReadParams(&model)

	if model.ID == 0 {
		rpcConn.GetRpcResponse().Error.Set(500, "id 不能为空")
		rpcConn.Write()
		return
	}

	err := basedboperat.Get(&model, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func unauthorizedMethod_list(rpcConn jsonrpc.JsonRpcConnection) {
	var model UnauthorizedMethod
	var models []UnauthorizedMethod
	var list basedboperat.List
	rpcConn.ReadParams(&list)

	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

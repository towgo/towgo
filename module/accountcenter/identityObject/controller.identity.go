package identityObject

import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
)

var method string = "/account/identity"

var defaultModel Identity
var defaultModels []Identity

func InitIdentityApi() {
	jsonrpc.SetFunc(method+"/create", create)
	jsonrpc.SetFunc(method+"/delete", delete)
	jsonrpc.SetFunc(method+"/update", update)
	jsonrpc.SetFunc(method+"/detail", detail)
	jsonrpc.SetFunc(method+"/list", list)
	jsonrpc.SetFunc(method+"/list_by_jurisdiction_code", listByJurisdictionCode)
	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.CREATE, method+"/create", "account_center", "创建角色"),
		dblog.NewOperateType(dblog.DELETE, method+"/delete", "account_center", "删除角色"),
		dblog.NewOperateType(dblog.UPDATE, method+"/update", "account_center", "修改角色"),
		dblog.NewOperateType(dblog.QUERY, method+"/detail", "account_center", "查询角色详情"),
		dblog.NewOperateType(dblog.QUERY, method+"/list", "account_center", "查询角色列表"),
		dblog.NewOperateType(dblog.QUERY, method+"/list_by_jurisdiction_code", "account_center", "查询角色码"),
	)
}

func create(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	_, err := model.Create()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func delete(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	count, err := model.Delete()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(count)
}

func update(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	err := model.Update()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func detail(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
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

func list(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	models := defaultModels
	var list basedboperat.List
	rpcConn.ReadParams(&list)
	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

func listByJurisdictionCode(rpcConn jsonrpc.JsonRpcConnection) {
	var params struct {
		Code string `json:"code"`
	}
	rpcConn.ReadParams(&params)
	if params.Code == "" {
		rpcConn.GetRpcResponse().Error.Set(500, "code 不能为空")
		rpcConn.Write()
		return
	}
	model := defaultModel
	models := defaultModels
	var list basedboperat.List
	list.And = map[string][]interface{}{"id": FindIdentityIdsByJurisdictionCode(params.Code)}
	list.Limit = -1
	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["rows"] = models
	rpcConn.WriteResult(result)
}

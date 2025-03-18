package apibilling

/*
用户账单计费模块管理API
by:liangliangit
*/

import (
	"context"
	"github.com/towgo/towgo/module/dblog"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
)

var method string = "/apibilling"
var defaultModel ApiBilling
var defaultModels []ApiBilling

func InitManageApi() {
	jsonrpc.SetFunc(method+"/create", create)
	jsonrpc.SetFunc(method+"/delete", delete)
	jsonrpc.SetFunc(method+"/update", update)
	jsonrpc.SetFunc(method+"/detail", detail)
	jsonrpc.SetFunc(method+"/list", list)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.CREATE, method+"/create", "account_center", "创建计费模块管理API"),
		dblog.NewOperateType(dblog.DELETE, method+"/delete", "account_center", "删除计费模块管理API"),
		dblog.NewOperateType(dblog.UPDATE, method+"/update", "account_center", "修改计费模块管理API"),
		dblog.NewOperateType(dblog.QUERY, method+"/detail", "account_center", "查看计费模块管理API详情"),
		dblog.NewOperateType(dblog.QUERY, method+"/list", "account_center", "查看计费模块管理API列表"),
	)
}

func create(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)

	dbSession, err := basedboperat.WithContext(context.Background())
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	dbSession.WithValue(jsonrpc.SESSION, rpcConn.GetRpcRequest().Session)

	_, err = dbSession.Create(&model)
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

	count, err := basedboperat.Delete(&model, nil, "id = ?", model.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(count)
}

func update(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	selectF := map[string]interface{}{}
	rpcConn.ReadParams(&model, &selectF)

	err := basedboperat.Update(&model, selectF, "id = ?", model.ID)
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

	findModel := defaultModel

	err := basedboperat.Get(&findModel, nil, "id = ?", model.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(findModel)
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

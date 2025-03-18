package systemconfig

import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
)

var defaultModel SystemConfig
var defaultModels []SystemConfig

func create(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	_, err := basedboperat.Create(&model)
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
	m := map[string]interface{}{}
	rpcConn.ReadParams(&model, &m)

	err := basedboperat.Update(&model, m, "id = ?", model.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
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

func api_systemconfig_get(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	rpcConn.WriteResult(model.Get(model.Key))

}
func api_systemconfig_set(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	model.Set(model.Key, model.Val)
}

package accessGateway

/*
统一接入网关 管理API
by:liangliangit
*/
import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
)

func init() {
	jsonrpc.SetFunc("/apiGateWay/create", api_route_group_create)
	jsonrpc.SetFunc("/apiGateWay/delete", api_route_group_delete)
	jsonrpc.SetFunc("/apiGateWay/update", api_route_group_update)
	jsonrpc.SetFunc("/apiGateWay/detail", api_route_group_detail)
	jsonrpc.SetFunc("/apiGateWay/list", api_route_group_list)
	jsonrpc.SetFunc("/apiGateWay/ApiRouteWay/list", api_route_group_api_route_way_list)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.CREATE, "/apiGateWay/create", "account_center", "创建网关路径"),
		dblog.NewOperateType(dblog.DELETE, "/apiGateWay/delete", "account_center", "删除网关路径"),
		dblog.NewOperateType(dblog.UPDATE, "/apiGateWay/update", "account_center", "修改网关路径"),
		dblog.NewOperateType(dblog.QUERY, "/apiGateWay/detail", "account_center", "查询网关路径详情"),
		dblog.NewOperateType(dblog.QUERY, "/apiGateWay/list", "account_center", "查询网关路径列表"),
		dblog.NewOperateType(dblog.QUERY, "/apiGateWay/ApiRouteWay/list", "account_center", "查询网关路径列表"),
	)

}

func api_route_group_create(rpcConn jsonrpc.JsonRpcConnection) {
	model := ApiRouteGroup{}
	rpcConn.ReadParams(&model)
	_, err := basedboperat.Create(&model)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func api_route_group_update(rpcConn jsonrpc.JsonRpcConnection) {
	model := ApiRouteGroup{}
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

func api_route_group_delete(rpcConn jsonrpc.JsonRpcConnection) {
	var model ApiRouteGroup
	rpcConn.ReadParams(&model)

	count, err := basedboperat.Delete(&model, model.ID, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(count)
}

func api_route_group_detail(rpcConn jsonrpc.JsonRpcConnection) {
	var model ApiRouteGroup
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

func api_route_group_list(rpcConn jsonrpc.JsonRpcConnection) {
	var model ApiRouteGroup
	var models []ApiRouteGroup
	var list basedboperat.List

	rpcConn.ReadParams(&list)
	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

func api_route_group_api_route_way_list(rpcConn jsonrpc.JsonRpcConnection) {
	var model ApiRouteWay
	var models []ApiRouteWay
	var list basedboperat.List
	rpcConn.ReadParams(&list)

	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

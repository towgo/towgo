package iolog

import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
)

func InitManageApi() {
	jsonrpc.SetFunc("/iolog/detail", IoLog_detail)
	jsonrpc.SetFunc("/iolog/list", IoLog_list)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.QUERY, "/iolog/detail", "account_center", "IO日志详情"),
		dblog.NewOperateType(dblog.QUERY, "/iolog/list", "account_center", "IO日志列表"),
	)
}

func IoLog_detail(rpcConn jsonrpc.JsonRpcConnection) {
	var model IoLog
	rpcConn.ReadParams(&model)

	if model.Guid == "" {
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

func IoLog_list(rpcConn jsonrpc.JsonRpcConnection) {
	var model IoLog
	var models []IoLog
	var list basedboperat.List
	rpcConn.ReadParams(&list)

	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

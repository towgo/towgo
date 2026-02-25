package licenseterminal

import (
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
)

var method string = "/license/terminal"

func init() {
	jsonrpc.SetFunc(method+"/list", list)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.QUERY, method+"/list", "account_center", "获取设备列表"),
	)
}

func list(rpcConn jsonrpc.JsonRpcConnection) {
	rpcConn.WriteResult(GetLicense())
}

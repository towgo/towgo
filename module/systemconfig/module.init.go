package systemconfig

import (
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
)

var method string = "/systemconfig"

func Init() {
	// jsonrpc.SetFunc("",)
	manage_api()
	operat_api()
	sync_db()
}

func sync_db() {
	xormDriver.Sync2(new(SystemConfig))
}

func manage_api() {
	jsonrpc.SetFunc(method+"/create", create)
	jsonrpc.SetFunc(method+"/delete", delete)
	jsonrpc.SetFunc(method+"/update", update)
	jsonrpc.SetFunc(method+"/list", list)
	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.CREATE, method+"/create", "account_center", "创建系统配置文件"),
		dblog.NewOperateType(dblog.DELETE, method+"/delete", "account_center", "删除系统配置文件"),
		dblog.NewOperateType(dblog.UPDATE, method+"/update", "account_center", "修改系统配置文件"),
		dblog.NewOperateType(dblog.QUERY, method+"/list", "account_center", "查看系统配置文件列表"),
	)
}

func operat_api() {
	jsonrpc.SetFunc(method+"/set", api_systemconfig_set)
	jsonrpc.SetFunc(method+"/get", api_systemconfig_get)

	dblog.BatchInsert(
		dblog.NewOperateType(dblog.UPLOAD, method+"/set", "account_center", "设置系统配置文件"),
		dblog.NewOperateType(dblog.QUERY, method+"/get", "account_center", "获取系统配置文件"),
	)
}

package ormDriver

import (
	"encoding/json"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/towgo/towgo/v2/dao/basedboperat"
	"github.com/towgo/towgo/v2/dao/ormDriver/gormDriver"
	"github.com/towgo/towgo/v2/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/v2/os/tcfg"
	"github.com/towgo/towgo/v2/os/tlog"
)

var dbCfg Config

func InitDatabase() {
	err := tcfg.GetConfig().LoadConfig()
	if err != nil {
		tlog.Error(gerror.Wrap(err, "database config init error"))
		return
	}
	err = tcfg.GetConfig().GetDataToStruct(ConfigNodeNameDatabase, &dbCfg)
	if err != nil {
		tlog.Error(gerror.Wrap(err, "database config init error"))
		return
	}
	if dbCfg.Mode == "" {
		dbCfg.Mode = "xorm"
	}
	nodes, err := json.Marshal(dbCfg.Nodes)
	if err != nil {
		panic(gerror.Wrap(err, "database config init error"))
	}
	switch dbCfg.Mode {
	case "xorm":
		var xormNodes []xormDriver.DsnConfig
		err = json.Unmarshal(nodes, &xormNodes)
		if err != nil {
			panic(gerror.Wrap(err, "database config init error"))
		}
		xormDriver.New(xormNodes)
		err = basedboperat.SetOrmEngine("xorm")
		if err != nil {
			panic(gerror.Wrap(err, "database config init error"))
		}
		break
	case "gorm":
		var gormNodes []gormDriver.DsnConfig
		err = json.Unmarshal(nodes, &gormNodes)
		if err != nil {
			panic(gerror.Wrap(err, "database config init error"))
		}
		gormDriver.New(gormNodes)
		err = basedboperat.SetOrmEngine("gorm")
		if err != nil {
			panic(gerror.Wrap(err, "database config init error"))
		}
		break
	}
}

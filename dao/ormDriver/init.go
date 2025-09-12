package ormDriver

import (
	"encoding/json"
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/gormDriver"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/errors/terror"
	"github.com/towgo/towgo/os/log"
	"github.com/towgo/towgo/os/tcfg"
)

var dbCfg Config

func InitDatabase() {
	err := tcfg.GetConfig().LoadConfig()
	if err != nil {
		log.Error(terror.Wrap(err, "database config init error"))
		return
	}
	err = tcfg.GetConfig().GetDataToStruct(ConfigNodeNameDatabase, &dbCfg)
	if err != nil {
		log.Error(terror.Wrap(err, "database config init error"))
		return
	}
	if dbCfg.Mode == "" {
		dbCfg.Mode = "xorm"
	}
	nodes, err := json.Marshal(dbCfg.Nodes)
	if err != nil {
		panic(terror.Wrap(err, "database config init error"))
	}
	switch dbCfg.Mode {
	case "xorm":
		var xormNodes []xormDriver.DsnConfig
		err = json.Unmarshal(nodes, &xormNodes)
		if err != nil {
			panic(terror.Wrap(err, "database config init error"))
		}
		xormDriver.New(xormNodes)
		err = basedboperat.SetOrmEngine("xorm")
		if err != nil {
			panic(terror.Wrap(err, "database config init error"))
		}
		break
	case "gorm":
		var gormNodes []gormDriver.DsnConfig
		err = json.Unmarshal(nodes, &gormNodes)
		if err != nil {
			panic(terror.Wrap(err, "database config init error"))
		}
		gormDriver.New(gormNodes)
		err = basedboperat.SetOrmEngine("gorm")
		if err != nil {
			panic(terror.Wrap(err, "database config init error"))
		}
		break
	}
}

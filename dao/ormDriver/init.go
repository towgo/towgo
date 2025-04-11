package ormDriver

import (
	"encoding/json"
	"github.com/towgo/towgo/dao/ormDriver/gormDriver"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/errors/terror"
	"github.com/towgo/towgo/os/tcfg"
)

var dbCfg Config

func init() {
	err := tcfg.GetConfig().GetDataToStruct(ConfigNodeNameDatabase, &dbCfg)
	if err != nil {
		panic(terror.Wrap(err, "database config init error"))
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
		break
	case "gorm":
		var gormNodes []gormDriver.DsnConfig
		err = json.Unmarshal(nodes, &gormNodes)
		if err != nil {
			panic(terror.Wrap(err, "database config init error"))
		}
		gormDriver.New(gormNodes)
		break
	}
}

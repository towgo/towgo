package main

import (
	"log"
	"net/http"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/moduledemos/accountcenter"
	"github.com/towgo/towgo/towgo"
)

func init() {
	//初始化xorm数据库驱动
	var xormDbConfigs []xormDriver.DsnConfig
	var configXorm xormDriver.DsnConfig
	configXorm.DbType = "mysql"
	configXorm.Dsn = "root:12345678@tcp(localhost:3306)/demo?charset=utf8mb4"
	configXorm.IsMaster = true
	xormDbConfigs = append(xormDbConfigs, configXorm)
	xormDriver.New(xormDbConfigs)

	//设定默认orm引擎
	err := basedboperat.SetOrmEngine("xorm")
	if err != nil {
		log.Print(err.Error())
	}

	xormDriver.Sync2(new(accountcenter.User))

}

func main() {

	towgo.SetFunc("/hello", hello)
	towgo.SetFunc("/login", login)

	towgo.NewCRUDJsonrpcAPI("/user", accountcenter.User{}, []accountcenter.User{}).RegAPI()

	tcpjsonrpcserver, err := towgo.NewTcpServer("0.0.0.0", "8090")
	if err != nil {
		log.Print(err)
	}
	tcpjsonrpcserver.Run()

	http.HandleFunc("/jsonrpc", towgo.HttpHandller)

	http.ListenAndServe("0.0.0.0:8080", nil)
}

func hello(rpcConn towgo.JsonRpcConnection) {
	var hello struct {
		Abc string
		Bcd bool
		Cfg int64
	}
	rpcConn.WriteResult(hello)
}

func login(rpcConn towgo.JsonRpcConnection) {
	rpcConn.WriteResult("ok logined")
}

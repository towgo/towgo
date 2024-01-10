package main

import (
	"log"
	"net/http"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/moduledemos/accountcenter"
	"github.com/towgo/towgo/towgo"
)

var appName string = "Account Center Module"
var appVersion string = "1.0.0"

var basePath = system.GetPathOfProgram()

type S string

var programPath string = system.GetPathOfProgram()
var config struct {
	DbType   string
	Dsn      string
	IsMaster bool
}

func init() {

	system.ScanConfigJson(programPath+"/config/config.json", &config)

	//初始化xorm数据库驱动
	var xormDbConfigs []xormDriver.DsnConfig
	var configXorm xormDriver.DsnConfig
	configXorm.DbType = config.DbType
	configXorm.Dsn = config.Dsn
	configXorm.IsMaster = config.IsMaster
	xormDbConfigs = append(xormDbConfigs, configXorm)
	xormDriver.New(xormDbConfigs)

	//设定默认orm引擎
	err := basedboperat.SetOrmEngine("xorm")
	if err != nil {
		log.Print(err.Error())
	}

}

func setu1(u accountcenter.User) {
	u.ID = 1
	u.Username = "abc"
}

func setu2(u *accountcenter.User) {
	u.ID = 1
	u.Username = "abc"
}

func setstrs(strs *[]string) {
	*strs = append(*strs, "a")
	*strs = append(*strs, "b")
	*strs = append(*strs, "c")
}

func setstrsarr(strs []string) {
	for k, v := range strs {
		if v == "c" {
			strs[k] = "c is find"
		}
	}
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

	//moduleClientInit()

	http.HandleFunc("/jsonrpc", towgo.HttpHandller)
	towgo.MethodToHttpPathInterface(http.DefaultServeMux)
	http.ListenAndServe("0.0.0.0:8080", nil)
}

func moduleClientInit() {
	var node towgo.EdgeServerNodeConfig
	system.ScanConfigJson(basePath+"config/togocdn.client.config.json", &node)
	node.Methods = towgo.GetMethods()
	node.ModuleName = appName
	for _, v := range node.ServerUrls {
		node.ServerUrl = v
		client := towgo.NewEdgeServerNode(node)
		client.Connect()
	}
}

func hello(rpcConn towgo.JsonRpcConnection) {
	var hello struct {
		Abc string
		Bcd bool
		Cfg int64
	}

	rpcConn.ReadParams(&hello)

	rpcConn.WriteResult(hello)
}

func login(rpcConn towgo.JsonRpcConnection) {
	rpcConn.WriteResult("ok logined")
}

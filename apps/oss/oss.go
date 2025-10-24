package main

import (
	"fmt"
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/processmanager"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/apibilling"
	"github.com/towgo/towgo/module/filestransfer"
	"log"
	"net/http"
	"os"
)

var appName string = "oss Module"
var appVersion string = "1.0.0"

var basePath = system.GetPathOfProgram()

func init() {
	//初始化xorm数据库驱动
	var dbconfig []xormDriver.DsnConfig
	system.ScanConfigJson(basePath+"/config/dbconfig.json", &dbconfig)
	xormDriver.New(dbconfig)

	//设定默认orm引擎
	err := basedboperat.SetOrmEngine("xorm")
	if err != nil {
		log.Print(err.Error())
	}

	apibilling.InitManageApi()
}
func main() {
	pm := processmanager.GetManager()
	for k, v := range os.Args {
		switch v {
		case "start":
			if k == 1 {
				if pm.Start() {
					log.Print("启动成功")
					start()
					return
				} else {
					log.Print("启动失败:" + pm.Error.Error())
					return
				}
			}
		case "restart":
			if k == 1 {
				if pm.ReStart() {
					log.Print("重启成功")
					start()
					return
				} else {
					log.Print("重启失败:" + pm.Error.Error())
				}
				return
			}
		case "stop":
			if k == 1 {
				if pm.Stop() {
					log.Print("程序停止成功")
				} else {
					log.Print("程序停止失败:程序没有运行")
				}
				return
			}
		case "version":
			if k == 1 {
				fmt.Print(appName + ":" + appVersion + "\n")
				os.Exit(0)
			}
			return
		}
	}
	log.Print("参数传递错误,有效参数如下:\n" + os.Args[0] + " start | stop | reload | stop")

}

func start() {
	config := struct {
		Serverport string `json:"serverport"`
	}{}

	system.ScanConfigJson(basePath+"/config/config.json", &config)
	moduleClientInit()
	filestransfer.SetApiHeader("/api")
	filestransfer.SetTokenKey("session")
	filestransfer.Auth(true)
	filestransfer.AllowCross(true)
	filestransfer.InitApi()

	http.ListenAndServe("0.0.0.0:"+config.Serverport, nil)

}

func moduleClientInit() {
	var node jsonrpc.EdgeServerNodeConfig
	system.ScanConfigJson(basePath+"config/togocdn.client.config.json", &node)
	node.Methods = jsonrpc.GetMethods()
	node.ModuleName = appName
	for _, v := range node.ServerUrls {
		node.ServerUrl = v
		client := jsonrpc.NewEdgeServerNode(node)
		client.Connect()
	}
}

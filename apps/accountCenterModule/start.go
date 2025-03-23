package main

import (
	"fmt"
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/gormDriver"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/module/accountcenter"
	"github.com/towgo/towgo/os/tcfg"
	"github.com/towgo/towgo/towgo"
	"log"
	"net/http"
	"os"

	// "github.com/towgo/towgo/lib/api"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/processmanager"
	"github.com/towgo/towgo/lib/www"
)

// var basePath = system.GetPathOfProgram()
var (
	appName    string = "Account Center Module"
	appVersion string = "1.0.0"

	conf *tcfg.Config = tcfg.GetConfig()
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// 初始化
func init() {

	//初始化xorm数据库驱动
	var xormDbConfig []xormDriver.DsnConfig
	conf.GetDataToStruct("database", &xormDbConfig)
	xormDriver.New(xormDbConfig)
	//初始化gorm数据库驱动
	var gormDbConfig []gormDriver.DsnConfig
	conf.GetDataToStruct("database", &gormDbConfig)
	gormDriver.New(gormDbConfig)

	//设定默认orm引擎
	err := basedboperat.SetOrmEngine("xorm")
	if err != nil {
		log.Print(err.Error())
	}

	basedboperat.SetGlobalCacheExpire(3000)

	// xormDriver.Sync2(new(User))

	// api.NewCRUDJsonrpcAPI("/user", User{}, []User{}).RegAPI()

	initJsonrpc()
}

// jsonrpc 初始化
func initJsonrpc() {

	//记录rpc日志

	//账户中心初始化
	accountcenter.InitManageApi()
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
	moduleClientInit()
	httpServer()
}

func httpServer() {
	port, err := conf.Get("server.port")
	if err != nil {
		return
	}
	http.HandleFunc("/jsonrpc", jsonrpc.HttpHandller)

	frontServer := www.WebServer{}
	frontServer.Wwwroot = "wwwroot"
	frontServer.Index = []string{"index.html"}
	towgo.HttpHandller()
	http.HandleFunc("/websocket/jsonrpc", jsonrpc.DefaultWebSocketServer.WebsocketServiceHandller.ServeHTTP)
	http.HandleFunc("/", frontServer.WebServerHandller)
	jsonrpc.MethodToHttpInterface(http.DefaultServeMux)
	log.Print("http服务运行中:0.0.0.0:" + port.(string) + "\n")
	http.ListenAndServe("0.0.0.0:"+port.(string), nil)

}

func moduleClientInit() {
	var node jsonrpc.EdgeServerNodeConfig
	err := conf.GetDataToStruct("togocdn.client", &node)
	if err != nil {
		panic(err)
		return
	}
	node.Methods = jsonrpc.GetMethods()
	node.ModuleName = appName
	for _, v := range node.ServerUrls {
		node.ServerUrl = v
		client := jsonrpc.NewEdgeServerNode(node)
		client.Connect()
	}
}

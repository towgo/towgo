package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// "github.com/towgo/towgo/cmd/utils"
	"github.com/google/uuid"
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/gormDriver"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/microsoftsso"
	"github.com/towgo/towgo/lib/processmanager"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/lib/www"
	"github.com/towgo/towgo/module/accountcenter"
	"github.com/towgo/towgo/module/dblog"
)

var appName string = "Account Center Module"
var appVersion string = "1.0.2.2024_12_24_15_58.wqy"

var basePath = system.GetPathOfProgram()

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// 初始化
func init() {

	//初始化xorm数据库驱动
	var dbconfig []xormDriver.DsnConfig
	system.ScanConfigJson(filepath.Join(basePath, "config/dbconfig.json"), &dbconfig)
	xormDriver.New(dbconfig)

	//初始化gorm数据库驱动
	var gormdbconfig []gormDriver.DsnConfig
	system.ScanConfigJson(filepath.Join(basePath, "config/dbconfig.json"), &gormdbconfig)
	gormDriver.New(gormdbconfig)

	//设定默认orm引擎
	err := basedboperat.SetOrmEngine("xorm")
	if err != nil {
		log.Print(err.Error())
	}

	// basedboperat.SetGlobalCacheExpire(3000)

	// xormDriver.Sync2(new(User))

	// api.NewCRUDJsonrpcAPI("/user", User{}, []User{}).RegAPI()

	initJsonrpc()
}
func GetUuid() string {
	u4 := uuid.New()
	uuid := strings.Split(u4.String(), "-")[4]
	return uuid
}

// jsonrpc 初始化
func initJsonrpc() {
	jsonrpc.SetSecretkey("FANHANINFOGOGOGO", "FANHANINFOGOGOGO")
	//记录rpc日志
	// uuid := GetUuid()
	/*jsonrpc.AfterExec = func(rpcConn jsonrpc.JsonRpcConnection) {
		method := rpcConn.GetRpcRequest().Method
		if sliceOperate.ListContainsAny(dblog.NotSaveLogMethod, method) {
			return
		}
		b, _ := json.Marshal(rpcConn.GetRpcResponse())
		// 去除密码关键字
		re := regexp.MustCompile(`,"password":"[^"]*"`)
		output := re.ReplaceAllString(rpcConn.Read(), "")

		paramResult := gjson.Get(output, "params")
		username := ""
		methodList := strings.Split(method, "/")
		if method != "/account/login" {
			account, _ := accountctx.Parse(rpcConn)
			if account != nil {
				username = account.Username
			}
		} else {
			usernameRes := gjson.Get(paramResult.Raw, "username")
			username = usernameRes.String()
		}
		dblog.Write("account_center", system.GetLocalIp(), methodList[1], "jsonrpc", paramResult.String(), string(b), rpcConn.GetRemoteAddr(), method, username)
	}*/
	//账户中心初始化
	accountcenter.InitManageApi()
	dblog.InitDbLogApi()
}

func main() {
	log.Println("start.....")
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

	conf := struct {
		Serverport string `json:"serverport"`
	}{}

	system.ScanConfigJson(filepath.Join(basePath, "config/config.json"), &conf)

	http.HandleFunc("/jsonrpc", jsonrpc.HttpHandller)

	var ssoconfig microsoftsso.SSOConfig

	system.ScanConfigJson(filepath.Join(basePath, "config/config.json"), &ssoconfig)
	microsoftsso.Init(ssoconfig)

	frontServer := www.WebServer{}
	frontServer.Wwwroot = "wwwroot"
	frontServer.Index = []string{"index.html"}

	http.HandleFunc("/websocket/jsonrpc", jsonrpc.DefaultWebSocketServer.WebsocketServiceHandller.ServeHTTP)
	http.HandleFunc("/", frontServer.WebServerHandller)
	log.Print("http服务运行中:0.0.0.0:" + conf.Serverport + "\n")
	http.ListenAndServe("0.0.0.0:"+conf.Serverport, nil)

}

func moduleClientInit() {
	var node jsonrpc.EdgeServerNodeConfig
	system.ScanConfigJson(filepath.Join(basePath, "config/togocdn.client.config.json"), &node)
	node.Methods = jsonrpc.GetMethods()
	node.ModuleName = appName
	for _, v := range node.ServerUrls {
		node.ServerUrl = v
		client := jsonrpc.NewEdgeServerNode(node)
		client.Connect()
	}
}

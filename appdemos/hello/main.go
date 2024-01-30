package main

import (
	"log"
	"net"
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

var writeQueue chan string = make(chan string, 1024)
var writeBlock chan int = make(chan int, 1)

func tcpReadHand(tcpClient net.Conn) {
	for {
		var buf []byte = make([]byte, 1024)
		tcpClient.Read(buf)

		//解析buf

		//如果是响应
		writeBlock <- 1
	}
}

func writeHand(tcpClient net.Conn) {
	for {
		s := <-writeQueue
		tcpClient.Write([]byte(s))
		<-writeBlock
	}
}

func main() {

	tcpClient, err := net.Dial("tcp", "192.168.1.1")

	go tcpReadHand(tcpClient)

	go writeHand(tcpClient)

	towgo.BeforExec = func(rpcConn towgo.JsonRpcConnection) {

	}

	towgo.AfterExec = func(rpcConn towgo.JsonRpcConnection) {

	}

	towgo.SetFunc("/hello", hello)
	towgo.SetFunc("/login", login)
	towgo.SetFunc("/create", create)

	towgo.NewCRUDJsonrpcAPI("/user", accountcenter.User{}, []accountcenter.User{}).RegAPI()

	tcpjsonrpcserver, err := towgo.NewTcpServer("0.0.0.0", "8090")
	if err != nil {
		log.Print(err)
	}
	tcpjsonrpcserver.Run()

	moduleClientInit()

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

}

func create(rpcConn towgo.JsonRpcConnection) {
	var m map[string]string
	m["abc"] = "a"
	var u accountcenter.User
	rpcConn.ReadParams(&u)
	basedboperat.Create(&u)
	rpcConn.WriteResult("ok")

}

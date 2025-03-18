package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/lib/www"
)

var basePath = system.GetPathOfProgram()
var appName string = "web front server"
var appVersion string = "1.0.0"

func main() {
	/*
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
	*/
	start()
}
func start() {
	conf := struct {
		Serverport string `json:"serverport"`
	}{}
	system.ScanConfigJson(filepath.Join(basePath, "config/config.json"), &conf)

	frontServer := www.WebServer{}
	frontServer.Wwwroot = "wwwroot"
	frontServer.Index = []string{"index.html"}

	http.HandleFunc("/", frontServer.WebServerHandller)
	downloadServer := www.WebServer{}
	downloadServer.Wwwroot = "wwwroot/download"
	http.HandleFunc("/download", downloadServer.WebServerHandller)

	log.Print("http服务运行中:0.0.0.0:" + conf.Serverport + "\n")
	http.ListenAndServe("0.0.0.0:"+conf.Serverport, nil)
}

package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/hello"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/order"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/user"
	"github.com/towgo/towgo/towgo"
)

var (
	version = "v1.0.0"
	app     = "towgo-gf-demo"
	server  = g.Server()
)

func init() {
	towgo.BindObject("/hello", hello.New())
	towgo.BindObject("/user", user.NewV1())
	towgo.BindObject("/order", order.NewV1())
	//http.HandleFunc("/jsonrpc", towgo.HttpHandller)
}

func start() {
	server.BindHandler("post:/jsonrpc", towgo.GhttpHandler)
	server.Run()

}

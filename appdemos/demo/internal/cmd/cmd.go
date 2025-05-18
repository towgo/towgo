package cmd

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/hello"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/order"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/user"
	"github.com/towgo/towgo/towgo"
	"net/http"
)

var (
	version = "v1.0.0"
	app     = "towgo-gf-demo"
)

func init() {
	towgo.BindObject("/hello", hello.New())
	towgo.BindObject("/user", user.NewV1())
	towgo.BindObject("/order", order.NewV1())
	http.HandleFunc("/jsonrpc", towgo.HttpHandller)
}

func start() error {
	ctx := context.Background()
	port, err := g.Config().Get(ctx, "server.port")
	if err != nil {
		return err
	}
	listenAddr := fmt.Sprintf("0.0.0.0:%s", port.String())
	glog.Infof(ctx, "启动成功 %+v", listenAddr)
	return http.ListenAndServe(listenAddr, nil)
}

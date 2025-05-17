package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/hello"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/order"
	"github.com/towgo/towgo/appdemos/demo/internal/controller/user"
	"github.com/towgo/towgo/appdemos/demo/internal/migrations"
	"github.com/towgo/towgo/lib/processmanager"
	"github.com/towgo/towgo/towgo"
	"net/http"

	"golang.org/x/net/context"
)

func init() {
	towgo.BindObject("/hello", hello.New())
	towgo.BindObject("/user", user.NewV1())
	towgo.BindObject("/order", order.NewV1())
	http.HandleFunc("/jsonrpc", towgo.HttpHandller)
}
func start() error {
	port, err := g.Config().Get(context.Background(), "server.port")
	if err != nil {
		return err
	}
	glog.Infof(context.Background(), "启动成功 %+v", port)
	return http.ListenAndServe("0.0.0.0:"+port.String(), nil)
}

var (
	version = "v1.0.0"
	Main    = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
	}
	DbSync = gcmd.Command{
		Name:  "sync",
		Usage: "DbSync 数据库迁移",
		Brief: " sync db info",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			migrations.Sync(ctx)
			return nil
		},
	}

	Start = gcmd.Command{
		Name:  "start",
		Usage: "start",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			pm := processmanager.GetManager()
			if pm.Start() {
				err = start()
			} else {
				err = pm.Error
			}
			return err
		},
	}

	Stop = gcmd.Command{
		Name:  "stop",
		Usage: "stop",
		Brief: "stop http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			pm := processmanager.GetManager()
			if pm.Stop() {
				glog.Infof(ctx, "停止成功")
			} else {
				glog.Infof(ctx, "停止失败")
			}

			return nil
		},
	}
	Restart = gcmd.Command{
		Name:  "restart",
		Usage: "restart",
		Brief: "restart http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			pm := processmanager.GetManager()
			if pm.ReStart() {
				glog.Infof(ctx, "重启成功")
				err = start()
			} else {
				glog.Infof(ctx, "停止失败")
			}
			return nil
		},
	}
	Version = gcmd.Command{
		Name:  "version",
		Usage: "version",
		Brief: "query version info",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			glog.Print(ctx, "version", g.Map{"app": version})
			return nil
		},
	}
)

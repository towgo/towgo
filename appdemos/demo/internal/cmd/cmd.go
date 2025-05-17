package cmd

import (
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/towgo/towgo/appdemos/demo/internal/migrations"
	"github.com/towgo/towgo/jsonrpcV2"
	"github.com/towgo/towgo/lib/processmanager"

	"golang.org/x/net/context"
)

var (
	Main = gcmd.Command{
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
				jsonrpcV2.NewJsonRpcServer().Start()
			} else {
				err = pm.Error
			}
			return err
		},
	}
)

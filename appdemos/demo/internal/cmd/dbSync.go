package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/appdemos/demo/internal/migrations"
)

var DbSync = gcmd.Command{
	Name:  "sync",
	Usage: "DbSync 数据库迁移",
	Brief: " sync db info 数据库迁移",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		err = migrations.Sync(ctx)
		if err != nil {
			return err
		}
		glog.Info(ctx, "db table sync deno!")
		return err
	},
}

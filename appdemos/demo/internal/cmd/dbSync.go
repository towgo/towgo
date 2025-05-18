package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/towgo/towgo/appdemos/demo/internal/migrations"
)

var DbSync = gcmd.Command{
	Name:  "sync",
	Usage: "DbSync 数据库迁移",
	Brief: " sync db info",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		migrations.Sync(ctx)
		return nil
	},
}

package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/appdemos/demo/internal/cmd"
)

func main() {
	ctx := gctx.GetInitCtx()
	root, err := gcmd.NewFromObject(cmd.Main)
	if err != nil {
		return
	}
	err = root.AddObject(cmd.DbSync, cmd.Start)
	if err != nil {
		glog.Error(ctx, err.Error())
	}
	root.Run(ctx)
}

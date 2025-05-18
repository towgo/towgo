package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/lib/processmanager"
)

var Stop = gcmd.Command{
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

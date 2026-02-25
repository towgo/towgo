package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/lib/processmanager"
)

var Restart = gcmd.Command{
	Name:  "restart",
	Usage: "restart",
	Brief: "restart http server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		pm := processmanager.GetManager()
		if pm.ReStart() {
			glog.Infof(ctx, "重启成功")
			err = server.Shutdown()
			if err != nil {
				return err
			}
			start()
		} else {
			glog.Infof(ctx, "停止失败")
		}
		return err
	},
}

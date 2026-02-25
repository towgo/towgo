package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/towgo/towgo/lib/processmanager"
)

var Start = gcmd.Command{
	Name:  "start",
	Usage: "start",
	Brief: "start http server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		pm := processmanager.GetManager()
		if pm.Start() {
			start()
		} else {
			err = pm.Error
		}
		return err
	},
}

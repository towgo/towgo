package cmd

import (
	"context"
	"github.com/gogf/gf/v2/os/gcmd"
	"log"
)

var Version = gcmd.Command{
	Name:  "version",
	Usage: "version",
	Brief: "query version info",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		log.Printf("\nversion info \napp : %+v \nversion :%+v ", app, version)
		return nil
	},
}

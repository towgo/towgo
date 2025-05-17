package cmd

import (
	"errors"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/towgo/towgo/appdemos/demo/internal/migrations"
	"github.com/towgo/towgo/towgo"
	"log"
	"net/http"

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
			towgo.BindObject("/hello", NewHello())
			log.Printf("%+v", towgo.GetMethods())
			http.HandleFunc("/jsonrpc", towgo.HttpHandller)
			err = http.ListenAndServe("0.0.0.0:8080", nil)

			return err
		},
	}
)

type Hello struct {
}

func NewHello() *Hello {
	return &Hello{}
}
func (h *Hello) Hello1(conn *towgo.Jsonrpcrequest) (rep *HelloRep, err error) {

	log.Println("Hello1 ---")
	return rep, err
}

func (h *Hello) Hello2(req *HelloReq) (rep *HelloRep, err error) {
	log.Printf("%+v", req)
	log.Println("Hello2 ---")
	rep = &HelloRep{
		Age:     30,
		Address: "Hello2",
	}
	return
}
func (h *Hello) Hello3(req *HelloReq) (rep *HelloRep, err error) {
	log.Println("Hello3 ---")
	log.Printf("%+v", req)
	log.Println("Hello3 ---")
	rep = &HelloRep{
		Age:     40,
		Address: "Hello3",
	}
	err = errors.New("Hello3")
	return
}
func (h *Hello) HelloName(req *HelloReq) (rep *HelloRep, err error) {
	log.Println("HelloName ---")
	log.Printf("%+v", req)
	log.Println("HelloName ---")
	rep = &HelloRep{
		Age:     40,
		Address: "HelloName",
	}
	err = errors.New("HelloName")
	return
}

type HelloReq struct {
	Id    uint   `p:"id"`
	Name  string `p:"name"`
	Email string `p:"email"`
}
type HelloRep struct {
	Age     uint   `json:"age"`
	Address string `json:"address"`
}

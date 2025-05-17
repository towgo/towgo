package hello

import (
	"context"
	"errors"
	v1 "github.com/towgo/towgo/appdemos/demo/api/hello/v1"
	"github.com/towgo/towgo/towgo"
	"log"
)

type Hello struct {
}

func New() *Hello {
	return &Hello{}
}
func (h *Hello) Hello1(conn towgo.JsonRpcConnection) {

	log.Println("Hello1 ---")
	conn.WriteResult("Hello1")
	return
}

func (h *Hello) Hello2(ctx context.Context, req *v1.HelloReq) (rep *v1.HelloRes, err error) {
	log.Printf("%+v", req)
	log.Println("Hello2 ---")
	rep = &v1.HelloRes{
		Age:     30,
		Address: "Hello2",
	}
	return
}
func (h *Hello) Hello3(ctx context.Context, req *v1.HelloReq) (rep *v1.HelloRes, err error) {
	log.Println("Hello3 ---")
	log.Printf("req = %+v", req)
	log.Println("Hello3 ---")
	rep = &v1.HelloRes{
		Age:     40,
		Address: "Hello3",
	}
	err = errors.New("Hello3")
	return
}
func (h *Hello) HelloName(ctx context.Context, req *v1.HelloReq) (rep *v1.HelloRes, err error) {
	log.Println("HelloName ---")
	log.Printf("%+v", req)
	log.Println("HelloName ---")
	rep = &v1.HelloRes{
		Age:     40,
		Address: "HelloName",
	}
	err = errors.New("HelloName")
	return
}

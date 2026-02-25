package v1

import "github.com/gogf/gf/v2/frame/g"

type HelloReq struct {
	g.Meta
	Id    uint   `p:"id" v:"required|min:1#id不能为空|id最小为1" d:"31"`
	Name  string `p:"name"`
	Email string `p:"email"`
}
type HelloRes struct {
	Age     uint   `json:"age"`
	Address string `json:"address"`
}

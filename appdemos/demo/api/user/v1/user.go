package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/towgo/towgo/appdemos/demo/internal/model"
	"github.com/towgo/towgo/appdemos/demo/internal/model/entity"
)

type GetListReq struct {
	g.Meta `path:"/user/list"`
	model.PageReq
}
type GetListRes struct {
	model.ListRes
}

type DetailReq struct {
	g.Meta `path:"/user/detail"`
	Id     int `p:"id"`
}
type DetailRes struct {
	*entity.User
}

type CreateReq struct {
	g.Meta `path:"/user/create"`
	entity.User
}
type CreateRes struct {
}

type UpdateReq struct {
	g.Meta `path:"/user/update"`
	entity.User
}
type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/user/delete"`
	Id     int `p:"id"  v:"required|min:1#id不能为空|id最小为1"`
}
type DeleteRes struct{}

package user

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/user/v1"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := c.userService.List(ctx, &req.PageReq)
	if err != nil {
		return
	}
	res = &v1.GetListRes{}
	res.ListRes = *list
	return
}

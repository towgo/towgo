package order

import (
	"context"
	"github.com/towgo/towgo/appdemos/demo/api/order/v1"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := c.orderServer.List(ctx, &req.PageReq)
	if err != nil {
		return
	}
	res = &v1.GetListRes{}
	res.ListRes = *list
	return
}

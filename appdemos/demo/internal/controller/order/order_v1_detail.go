package order

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/order/v1"
)

func (c *ControllerV1) Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error) {
	detail, err := c.orderServer.Detail(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &v1.DetailRes{
		detail,
	}
	return
}

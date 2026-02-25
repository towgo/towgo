package order

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/order/v1"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.orderServer.Update(ctx, &req.Order)
	return
}

package order

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/order/v1"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	err = c.orderServer.Create(ctx, &req.Order)
	return
}

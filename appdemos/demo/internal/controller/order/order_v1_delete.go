package order

import (
	"context"

	"github.com/towgo/towgo/v2/appdemos/demo/api/order/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.orderServer.Delete(ctx, req.Id)
	return
}

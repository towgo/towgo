package user

import (
	"context"

	"github.com/towgo/towgo/v2/appdemos/demo/api/user/v1"
)

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = c.userService.Delete(ctx, req.Id)
	return
}

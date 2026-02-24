package user

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/user/v1"
)

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = c.userService.Update(ctx, &req.User)
	return
}

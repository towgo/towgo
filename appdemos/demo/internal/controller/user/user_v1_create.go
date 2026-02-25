package user

import (
	"context"

	"github.com/towgo/towgo/appdemos/demo/api/user/v1"
)

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	err = c.userService.Create(ctx, &req.User)
	return
}

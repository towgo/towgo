package user

import (
	"context"
	"github.com/towgo/towgo/appdemos/demo/internal/dao"
	"github.com/towgo/towgo/appdemos/demo/internal/model"
	"github.com/towgo/towgo/appdemos/demo/internal/model/entity"
)

type UserService struct{}

func New() *UserService {
	return &UserService{}
}
func (userService *UserService) List(ctx context.Context, req *model.PageReq) (res *model.ListRes, err error) {
	var (
		clr   = dao.User.Columns()
		orm   = dao.User.Ctx(ctx)
		count = 0
		list  []entity.User
	)
	err = orm.Page(req.PageNum, req.PageSize).Order(clr.Id+" desc").ScanAndCount(&list, &count, false)
	if err != nil {
		return
	}
	res = &model.ListRes{}
	res.Rows = list
	res.Count = count
	if err != nil {
		return
	}
	return
}

func (userService *UserService) Detail(ctx context.Context, id int) (u *entity.User, err error) {
	err = dao.User.Ctx(ctx).WherePri(id).Scan(&u)
	return
}

func (userService *UserService) Create(ctx context.Context, u *entity.User) (err error) {
	_, err = dao.User.Ctx(ctx).Insert(u)
	return
}

func (userService *UserService) Update(ctx context.Context, u *entity.User) (err error) {
	_, err = dao.User.Ctx(ctx).WherePri(u.Id).Update(u)
	return
}

func (userService *UserService) Delete(ctx context.Context, id int) (err error) {
	_, err = dao.User.Ctx(ctx).WherePri(id).Delete()
	return
}

package order

import (
	"context"
	"github.com/towgo/towgo/appdemos/demo/internal/dao"
	"github.com/towgo/towgo/appdemos/demo/internal/model"
	"github.com/towgo/towgo/appdemos/demo/internal/model/entity"
)

type OrderService struct{}

func New() *OrderService {
	return &OrderService{}
}
func (orderService *OrderService) List(ctx context.Context, req *model.PageReq) (res *model.ListRes, err error) {
	var (
		clr   = dao.Order.Columns()
		orm   = dao.Order.Ctx(ctx)
		count = 0
		list  []entity.Order
	)

	err = orm.Page(req.PageNum, req.PageSize).Order(clr.Id+" desc").ScanAndCount(&list, &count, false)
	res = &model.ListRes{}
	res.Rows = list
	res.Count = count
	if err != nil {
		return
	}
	return
}

func (orderService *OrderService) Detail(ctx context.Context, id int) (order *entity.Order, err error) {
	err = dao.Order.Ctx(ctx).WherePri(id).Scan(&order)
	return
}

func (orderService *OrderService) Create(ctx context.Context, order *entity.Order) (err error) {
	_, err = dao.Order.Ctx(ctx).Insert(order)
	return
}

func (orderService *OrderService) Update(ctx context.Context, order *entity.Order) (err error) {
	_, err = dao.Order.Ctx(ctx).WherePri(order.Id).Update(order)
	return
}

func (orderService *OrderService) Delete(ctx context.Context, id int) (err error) {
	_, err = dao.Order.Ctx(ctx).WherePri(id).Delete()
	return
}

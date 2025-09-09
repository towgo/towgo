// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderDao is the data access object for table order.
type OrderDao struct {
	table   string       // table is the underlying table name of the DAO.
	group   string       // group is the database configuration group name of current DAO.
	columns OrderColumns // columns contains all the column names of Table for convenient usage.
}

// OrderColumns defines and stores column names for table order.
type OrderColumns struct {
	Id           string //
	OrderNo      string //
	CustomerName string //
	Amount       string //
	Status       string //
	PayTime      string //
	CreateTime   string //
	UpdateTime   string //
}

// orderColumns holds the columns for table order.
var orderColumns = OrderColumns{
	Id:           "id",
	OrderNo:      "order_no",
	CustomerName: "customer_name",
	Amount:       "amount",
	Status:       "status",
	PayTime:      "pay_time",
	CreateTime:   "create_time",
	UpdateTime:   "update_time",
}

// NewOrderDao creates and returns a new DAO object for table data access.
func NewOrderDao() *OrderDao {
	return &OrderDao{
		group:   "default",
		table:   "order",
		columns: orderColumns,
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *OrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current dao.
func (dao *OrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of current dao.
func (dao *OrderDao) Columns() OrderColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current dao.
func (dao *OrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *OrderDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *OrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

package migrations

import (
	"time"
)

func init() {
	AddMigrate(new(Order))
}
func (Order) TableName() string {
	return "order"
}

type Order struct {
	Id           uint      `json:"id"           orm:"id"            description:""` //
	OrderNo      string    `json:"orderNo"      orm:"order_no"      description:""` //
	CustomerName string    `json:"customerName" orm:"customer_name" description:""` //
	Amount       float64   `json:"amount"       orm:"amount"        description:""` //
	Status       string    `json:"status"       orm:"status"        description:""` //
	PayTime      time.Time `json:"payTime"      orm:"pay_time"      description:""` //
	CreateTime   time.Time `json:"createTime"   orm:"create_time"   description:""` //
	UpdateTime   time.Time `json:"updateTime"   orm:"update_time"   description:""` //
}

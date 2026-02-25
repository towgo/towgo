// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Order is the golang structure for table order.
type Order struct {
	Id           uint64      `json:"id"           orm:"id"            description:""` //
	OrderNo      string      `json:"orderNo"      orm:"order_no"      description:""` //
	CustomerName string      `json:"customerName" orm:"customer_name" description:""` //
	Amount       float64     `json:"amount"       orm:"amount"        description:""` //
	Status       string      `json:"status"       orm:"status"        description:""` //
	PayTime      *gtime.Time `json:"payTime"      orm:"pay_time"      description:""` //
	CreateTime   *gtime.Time `json:"createTime"   orm:"create_time"   description:""` //
	UpdateTime   *gtime.Time `json:"updateTime"   orm:"update_time"   description:""` //
}

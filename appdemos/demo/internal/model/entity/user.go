// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User is the golang structure for table user.
type User struct {
	Id           uint64      `json:"id"           orm:"id"            description:""` //
	Name         string      `json:"name"         orm:"name"          description:""` //
	Email        string      `json:"email"        orm:"email"         description:""` //
	Age          uint        `json:"age"          orm:"age"           description:""` //
	Birthday     *gtime.Time `json:"birthday"     orm:"birthday"      description:""` //
	MemberNumber string      `json:"memberNumber" orm:"member_number" description:""` //
	ActivatedAt  *gtime.Time `json:"activatedAt"  orm:"activated_at"  description:""` //
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""` //
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:""` //
}

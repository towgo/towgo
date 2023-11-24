package models

type BaseModel struct {
	ID        int64 `json:"id"`
	CreatedAt int64 `json:"created_at" xorm:"created comment('创建时间')"`
	UpdatedAt int64 `json:"updated_at" xorm:"updated comment('修改时间')"`
}
type IDModel struct {
	ID int64 `json:"id"`
}
type CreatedAndUpdatedAt struct {
	CreatedAt int64 `json:"created_at" xorm:"created comment('创建时间')"`
	UpdatedAt int64 `json:"updated_at" xorm:"updated comment('修改时间')"`
}

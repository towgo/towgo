package accountcenter

import "github.com/towgo/towgo/dao/basedboperat"

func (Userinfo) TableName() string {
	return "userinfo"
}

type Userinfo struct {
	ID            int64           `json:"id"`
	Uid           int64           `json:"uid"`
	Nickname      string          `json:"nickname"`
	Address       string          `json:"address"`
	Userorderinfo []Userorderinfo `json:"userorderinfo" xorm:"-"`
}

func (u *Userinfo) AfterQuery(session basedboperat.DbTransactionSession) error {

	var userorderinfo Userorderinfo
	var userorderinfos []Userorderinfo
	var list basedboperat.List
	list.Limit = -1

	session.ListScan(&list, &userorderinfo, &userorderinfos)

	u.Userorderinfo = userorderinfos
	return nil
}

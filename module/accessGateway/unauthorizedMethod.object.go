package accessGateway

import (
	"errors"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
)

/*
未授权可访问的API
*/

func init() {
	xormDriver.Sync2(new(UnauthorizedMethod))
}

func (UnauthorizedMethod) TableName() string {
	return "unauthorized_method"
}

type UnauthorizedMethod struct {
	ID       int64  `json:"id"`
	Method   string `json:"method"`
	Type     string `json:"type"`
	IsSystem bool   `json:"is_system"`
}

func (u *UnauthorizedMethod) BeforDelete() error {
	var findModel UnauthorizedMethod
	err := basedboperat.Get(&findModel, nil, "id = ?", u.ID)
	if err != nil {
		return err
	}
	if findModel.ID == 0 {
		return errors.New("记录不存在")
	}
	if findModel.IsSystem {
		return errors.New("无法删除系统接口")
	}
	return nil
}

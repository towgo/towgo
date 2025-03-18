package systemconfig

import (
	"errors"

	"github.com/towgo/towgo/dao/basedboperat"
)

func (SystemConfig) TableName() string {
	return "system_config"
}

type SystemConfig struct {
	ID       int64  `json:"id"`
	Key      string `json:"key" xorm:"unique"`
	Val      string `json:"val"`
	Name     string `json:"name"`
	Describe string `json:"describe"`
}

func (sc *SystemConfig) Get(Key string) string {
	var sctmp SystemConfig
	basedboperat.Get(&sctmp, nil, "key = ?", Key)
	return sctmp.Val
}

func (sc *SystemConfig) Set(key, val string) error {
	if key == "" {
		return errors.New("key 不能为空")
	}
	var sctmp SystemConfig
	sctmp.Val = val
	return basedboperat.Update(&sctmp, "val", "key = ?", key)
}

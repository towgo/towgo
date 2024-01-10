package accountcenter

import (
	"errors"
	"log"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/accountcenter/accountctx"
	"github.com/towgo/towgo/towgo"
)

func (u *User) InputCheck(session basedboperat.DbTransactionSession) error {

	if u.Username == "" {
		return errors.New("用户名不能为空")
	}

	if u.Password == "" {
		return errors.New("密码不能为空")
	}

	return nil
}

func (u *User) BeforeCreate(session basedboperat.DbTransactionSession) error {

	salt := system.RandChar(6)
	u.Salt = salt
	u.Password = system.MD5(system.MD5(u.Password) + salt)

	var k towgo.ContextKey = towgo.JSON_RPC_CONNECTION_CONTEXT_KEY
	connInterface := session.Value(k)
	rpcConn := connInterface.(towgo.JsonRpcConnection)
	account, err := accountctx.Parse(rpcConn)
	if err != nil {
		return err
	}
	log.Print(account.ID)

	return nil
}

func (u *User) AfterQuery(session basedboperat.DbTransactionSession) error {
	var userinfo Userinfo
	err := session.Get(&userinfo, nil, "uid = ?", u.ID)
	if err != nil {
		return err
	}
	u.Userinfo = userinfo
	return nil
}

package accountctx

import (
	"encoding/json"
	"errors"

	"github.com/towgo/towgo/lib/jsonrpc"
)

type ctxKey string

const (
	CTXKEY ctxKey = "accountcenter/accountctx/account"
)

type Account struct {
	ID            int64          `json:"id"`
	Username      string         `json:"username"`
	Nickname      string         `json:"nickname"`
	Password      string         `json:"password"`
	Salt          string         `json:"-"` //密码加盐
	Email         string         `json:"email"`
	CanDelete     bool           `json:"-"`
	Departments   []Department   `json:"departments" gorm:"-" xorm:"-"`   //所属部门
	Identitys     []Identity     `json:"identitys" gorm:"-" xorm:"-"`     //所属身份
	Jurisdictions []Jurisdiction `json:"jurisdictions" gorm:"-" xorm:"-"` //(账户拥有的权限)
	Token         string         `json:"token" gorm:"-" xorm:"-"`
	CreatedAt     int64          `json:"created_at"` //创建时间
	UpdatedAt     int64          `json:"updated_at"` //更新时间
}

type Department struct {
	ID            int64             `json:"id"`  //部门ID
	Fid           int64             `json:"fid"` //上级部门
	Parent        *ParentDepartment `json:"parent" gorm:"-" xorm:"-"`
	Name          string            `json:"name"`                            //部门名称唯一key
	Nickname      string            `json:"nickname"`                        //部门名称
	Jurisdictions []int64           `json:"jurisdictions" gorm:"-" xorm:"-"` //部门权限
	Remark        string            `json:"remark"`                          //备注
	UpdatedAt     int64             `json:"updated_at"`
	CreatedAt     int64             `json:"created_at"`
}

type ParentDepartment struct {
	ID            int64   `json:"id"`                              //部门ID
	Name          string  `json:"name"`                            //部门名称唯一key
	Nickname      string  `json:"nickname"`                        //部门名称
	Jurisdictions []int64 `json:"jurisdictions" gorm:"-" xorm:"-"` //部门权限
	Remark        string  `json:"remark"`                          //备注
	UpdatedAt     int64   `json:"updated_at"`
	CreatedAt     int64   `json:"created_at"`
}

type Jurisdiction struct {
	ID   int64  `json:"id"`
	Fid  int64  `json:"fid"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Jurisdictions struct {
	ID     int64           `json:"id"`
	Fid    int64           `json:"fid"`
	Code   string          `json:"code"`
	Name   string          `json:"name"`
	Childs []Jurisdictions `json:"childs" xorm:"-" gorm:"-"`
}

type Identity struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`      //唯一编码
	Name      string `json:"name"`      //显示名称
	Remark    string `json:"remark"`    //显示名称
	IsSystem  bool   `json:"is_system"` //是否为系统内置身份
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`

	Accounts      []int64 `json:"accounts" gorm:"-" xorm:"-"`
	Jurisdictions []int64 `json:"jurisdictions" gorm:"-" xorm:"-"`
}

func Parse(rpcConn jsonrpc.JsonRpcConnection) (*Account, error) {
	var account Account
	ctx := rpcConn.GetRpcRequest().Ctx
	accountInterface, ok := ctx["account"]
	if !ok {
		if rpcConn.GetRpcRequest().Session == "" {
			return nil, errors.New("没有账号信息")
		}

		err := jsonrpc.Call("/account/myinfo", rpcConn.GetRpcRequest().Session, nil, &account)
		if err != nil {
			return nil, err
		}
		return &account, nil
	}

	b, err := json.Marshal(accountInterface)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (a *Account) HasIdentity(identityID int64) bool {
	if a.HasJurisdiction("superadmin") {
		return true
	}

	for _, v := range a.Identitys {
		if v.ID == identityID {
			return true
		}
	}
	return false
}

func (a *Account) HasJurisdiction(jurisdictionCode string) bool {
	for _, v := range a.Jurisdictions {
		if v.Code == jurisdictionCode {
			return true
		}
	}
	return false
}

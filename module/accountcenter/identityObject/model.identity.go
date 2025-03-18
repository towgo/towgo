package identityObject

import (
	"errors"

	"github.com/towgo/towgo/dao/basedboperat"
)

//身份管理器

func (Identity) TableName() string {
	return "identitys"
}

type Identity struct {
	ID        int64  `json:"id"`
	Classify  string `json:"classify"`  //分类标记
	Code      string `json:"code"`      //唯一编码
	Name      string `json:"name"`      //显示名称
	Remark    string `json:"remark"`    //显示名称
	IsSystem  bool   `json:"is_system"` //是否为系统内置身份
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`

	Navs          []int64 `json:"navs" gorm:"-" xorm:"-"`
	Accounts      []int64 `json:"accounts" gorm:"-" xorm:"-"`
	Jurisdictions []int64 `json:"jurisdictions" gorm:"-" xorm:"-"`
}

// 更新字段（空字段也会更新）
func (Identity) UpdateFields() (fields []string) {
	fields = []string{
		"code",
		"name",
		"remark",
	}
	return
}

func (IdentityNavs) TableName() string {
	return "identitys_navs"
}

type IdentityNavs struct {
	ID         int64
	IdentityId int64 `json:"identity_id"`
	NavID      int64 `json:"nav_id"`
}

func (IdentityAccounts) TableName() string {
	return "identitys_accounts"
}

type IdentityAccounts struct {
	ID         int64
	IdentityId int64 `json:"identity_id"`
	AccountId  int64 `json:"account_id"`
}

func (IdentityJurisdictions) TableName() string {
	return "identitys_jurisdictions"
}

type IdentityJurisdictions struct {
	ID             int64
	IdentityId     int64 `json:"identity_id"`
	JurisdictionId int64 `json:"jurisdiction_id"`
}

func (i *Identity) Create() (int64, error) {
	i.ID = 0
	i.IsSystem = false
	rowsAffected, err := basedboperat.Create(i)
	if err != nil {
		return rowsAffected, err
	}
	if i.ID <= 0 {
		return 0, errors.New("identity id not be 0")
	}
	rowsAffected2, err := i.CreateRelation()
	rowsAffected = rowsAffected + rowsAffected2
	return rowsAffected, err
}

func (i *Identity) Update() error {
	var find_i Identity
	basedboperat.Get(&find_i, nil, "id = ?", i.ID)
	if find_i.ID <= 0 {
		return errors.New("记录不存在")
	}
	i.DeleteRelationForUpdate()
	i.CreateRelationForUpdate()
	basedboperat.Update(i, i.UpdateFields(), "id = ?", i.ID)
	return nil
}

func (i *Identity) Delete() (int64, error) {
	var find_i Identity
	basedboperat.Get(&find_i, nil, "id = ?", i.ID)
	if find_i.IsSystem {
		return 0, errors.New("无法删除系统内置角色")
	}
	i.DeleteRelation()
	return basedboperat.Delete(i, i.ID, nil)
}

// 删除关联数据
func (i *Identity) DeleteRelation() {

	var ia IdentityAccounts
	var ij IdentityJurisdictions
	var in IdentityNavs
	basedboperat.SqlExec("delete from "+ia.TableName()+" where identity_id = ?", i.ID)
	basedboperat.SqlExec("delete from "+ij.TableName()+" where identity_id = ?", i.ID)
	basedboperat.SqlExec("delete from "+in.TableName()+" where identity_id = ?", i.ID)

}

// 更新时删除关联数据
func (i *Identity) DeleteRelationForUpdate() {
	var ij IdentityJurisdictions
	var in IdentityNavs
	basedboperat.SqlExec("delete from "+ij.TableName()+" where identity_id = ?", i.ID)
	basedboperat.SqlExec("delete from "+in.TableName()+" where identity_id = ?", i.ID)
}

// 创建关联数据
func (i *Identity) CreateRelation() (int64, error) {
	var rowsAffected int64

	for _, v := range i.Navs {
		var in IdentityNavs
		in.NavID = v
		in.IdentityId = i.ID
		rowsAffected2, _ := basedboperat.Create(&in)
		rowsAffected = rowsAffected + rowsAffected2
	}

	for _, v := range i.Jurisdictions {
		var j IdentityJurisdictions
		j.JurisdictionId = v
		j.IdentityId = i.ID
		rowsAffected2, _ := basedboperat.Create(&j)
		rowsAffected = rowsAffected + rowsAffected2
	}
	return rowsAffected, nil
}

// 创建关联数据
func (i *Identity) CreateRelationForUpdate() (int64, error) {
	var rowsAffected int64
	for _, v := range i.Navs {
		var in IdentityNavs
		in.NavID = v
		in.IdentityId = i.ID
		rowsAffected2, _ := basedboperat.Create(&in)
		rowsAffected = rowsAffected + rowsAffected2
	}

	for _, v := range i.Jurisdictions {
		var j IdentityJurisdictions
		j.JurisdictionId = v
		j.IdentityId = i.ID
		rowsAffected2, _ := basedboperat.Create(&j)
		rowsAffected = rowsAffected + rowsAffected2
	}
	return rowsAffected, nil
}

func (i *Identity) AfterQuery() {
	var in IdentityNavs
	var ins []IdentityNavs
	var ij IdentityJurisdictions
	var ijs []IdentityJurisdictions

	basedboperat.SqlQueryScan(&ins, "select * from "+in.TableName()+" where identity_id = ?", i.ID)
	basedboperat.SqlQueryScan(&ijs, "select * from "+ij.TableName()+" where identity_id = ?", i.ID)

	for _, v := range ins {
		i.Navs = append(i.Navs, v.NavID)
	}

	for _, v := range ijs {
		i.Jurisdictions = append(i.Jurisdictions, v.JurisdictionId)
	}

}

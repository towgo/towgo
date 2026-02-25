package accountcenter

import (
	"crypto/tls"
	"errors"
	"fmt"
	// "log"

	"github.com/go-ldap/ldap/v3"
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/accountcenter/identityObject"
	"github.com/towgo/towgo/module/accountcenter/jurisdictionObject"
	"github.com/towgo/towgo/utils"
)

func (Account) TableName() string {
	return "accounts"
}

// func (*Account) CacheExpire() int64 {
// 	return 5000
// }

// 账户对象 关联账户信息
type Account struct {
	ID                    int64                             `json:"id"`
	MsUid                 string                            `json:"ms_uid"`
	Username              string                            `json:"username" xorm:"unique"`
	Nickname              string                            `json:"nickname"`
	Password              string                            `json:"password"`
	Salt                  string                            `json:"-"` //密码加盐
	Email                 string                            `json:"email"`
	CanDelete             bool                              `json:"-"`
	Sso                   bool                              `json:"sso"`
	AccessToken           string                            `json:"-"`
	WeixinUserid          string                            `json:"weixin_userid"`                   //企业微信userid
	WeComId               string                            `json:"we_com_id"`                       //企业微信userid
	Mobile                string                            `json:"mobile"`                          //手机号
	Departments           []Department                      `json:"departments" gorm:"-" xorm:"-"`   //所属部门
	Identitys             []identityObject.Identity         `json:"identitys" gorm:"-" xorm:"-"`     //所属身份
	Jurisdictions         []jurisdictionObject.Jurisdiction `json:"jurisdictions" gorm:"-" xorm:"-"` //(账户拥有的权限)
	Token                 string                            `json:"token" gorm:"-" xorm:"-"`
	ThirdDepartmentId     string                            `json:"third_department_id" gorm:"-" xorm:"-"`
	UserToken             *UserToken                        `json:"-" gorm:"-" xorm:"-"`
	AccountImg            string                            `json:"account_img" xorm:"account_img"`
	AccountActivateStatus string                            `json:"account_activate_status" gorm:"-" xorm:"-"`
	Groups                []Group                           `json:"groups" gorm:"-" xorm:"-"`
	CreatedAt             int64                             `json:"created_at" xorm:"created"` //创建时间
	UpdatedAt             int64                             `json:"updated_at" xorm:"updated"` //更新时间
}

// 查看当前账户对象是否拥有权限
func (a *Account) HasJurisdiction(jurisdictionCode string) bool {
	for _, v := range a.Jurisdictions {
		if v.Code == jurisdictionCode {
			return true
		}
	}
	return false
}

type AccountActivateStatus struct {
	ID        int64  `json:"id"`
	AccountId int64  `json:"account_id"`
	Status    string `json:"status" xorm:"varchar(125)"` //是否激活 activate disable
	CreatedAt int64  `json:"created_at"`                 //创建时间
	UpdatedAt int64  `json:"updated_at"`                 //更新时间
}

func (AccountActivateStatus) TableName() string {
	return "account_activate_status"
}

type Group struct {
	ID        int64  `json:"id"`
	Name      string `json:"name" xorm:"unique varchar(256)"` //组名
	CreatedAt int64  `json:"created_at"`                      //创建时间
	UpdatedAt int64  `json:"updated_at"`                      //更新时间
}

func (Group) TableName() string {
	return "group"
}

type AccountGroupRelation struct {
	ID        int64 `json:"id"`
	AccountId int64 `json:"account_id"`
	GroupId   int64 `json:"group_id"`
	CreatedAt int64 `json:"created_at"` //创建时间
	UpdatedAt int64 `json:"updated_at"` //更新时间
}

func (AccountGroupRelation) TableName() string {
	return "account_group_relation"
}

// 身份查询()
func (a *Account) QueryIdentity(code string) bool {

	var ia identityObject.IdentityAccounts
	var ias []identityObject.IdentityAccounts
	ia_list := basedboperat.List{}
	ia_list.Limit = -1
	ia_list.And = map[string][]interface{}{
		"account_id": []interface{}{a.ID},
	}
	basedboperat.ListScan(&ia_list, ia, &ias)

	var id identityObject.Identity
	var ids []identityObject.Identity
	id_list := basedboperat.List{}
	id_list.Limit = -1
	id_list.Field = []string{"id"}
	id_c := []interface{}{}
	for _, v := range ias {
		id_c = append(id_c, v.IdentityId)
	}

	id_list.And = map[string][]interface{}{
		"code": []interface{}{code},
		"id":   id_c,
	}

	basedboperat.ListScan(&id_list, id, &ids)
	return id_list.Count > 0
}

type GetUserIdByPhoneResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Userid  string `json:"userid"`
}

// /get/qiyeweixin/user_id
func GetQiYeWeiXinUserid(request *jsonrpc.Jsonrpcrequest, phone string) (GetUserIdByPhoneResponse, error) {
	//
	requestP := map[string]string{"mobile": phone}
	var getUserIdByPhoneResponse GetUserIdByPhoneResponse
	err := jsonrpc.CallGateWay("/get/qiyeweixin/user_id", request.Session, requestP, &getUserIdByPhoneResponse)
	if err != nil {
		return getUserIdByPhoneResponse, err
	}
	return getUserIdByPhoneResponse, nil
}

// 注册
func (a *Account) Reg(username, password string) error {

	e := a.CheckForInput(username, password)
	if e != nil {
		return e
	}

	//数据库查询出用户信息

	findAccount := Account{}
	basedboperat.Get(&findAccount, nil, "username = ?", username)

	//检查用户名是否存在
	if username == findAccount.Username {
		return errors.New("用户名已存在")
	}

	//生成密码
	a.NewPassword(password)

	a.Username = username
	a.CanDelete = true

	_, err := basedboperat.Create(a) // 通过数据的指针来创建
	if err != nil {
		return err
	}
	_, err = a.CreateRelation()
	return err
}

// 检查输入参数
func (a *Account) CheckForInput(username, password string) error {

	if username == "" {
		return errors.New("用户名不能为空")
	}

	if password == "" {
		return errors.New("密码不能为空")
	}

	//防sql注入
	if system.FilteredSQLInject(username) {
		return errors.New("用户名存在系统保留或非法的字符")
	}
	return nil
}

func (a *Account) NewPassword(newpassword string) {
	if newpassword == "" {
		return
	}
	//加密密码

	password := system.SHA1(newpassword)

	//生成salt
	salt := system.RandCharCrypto(6)

	//密码加盐
	password = password + salt

	//混合加密
	password = system.SHA1(password)

	a.Password = password
	a.Salt = salt
}

// 用户登录
func (a *Account) Login(username, password string) error {

	erro := a.CheckForInput(username, password)
	if erro != nil {
		return erro
	}

	//通过用户名查询用户数据
	err := basedboperat.Get(a, nil, "username = ?", username)

	if err != nil {
		return err
	}
	//检查用户名是否存在

	//判断用户是否存在
	if a.Username == "" {
		return errors.New("用户名不存在")
	}
	var accountActivateStatus AccountActivateStatus
	basedboperat.Get(&accountActivateStatus, nil, "account_id = ?", a.ID)
	if accountActivateStatus.Status == "disable" {
		return errors.New("当前账号已冻结")
	}
	//加密密码
	upassword := system.SHA1(password)

	//撒盐
	upassword = upassword + a.Salt

	//混合加密
	upassword = system.SHA1(upassword)

	//判断密码是否一致
	if a.Password != upassword {
		//不一致:返回错误
		return errors.New("密码错误")
	}

	//验证通过

	//生成用户信息
	a.UserToken = NewToken(a)

	return nil
}

// 用户登录
func (a *Account) MicrosoftAdLogin(username, password string) error {
	if username == "" {
		return errors.New("用户名或密码不能为空")
	}

	if password == "" {
		return errors.New("用户名或密码不能为空")
	}
	microConfig := utils.GetMicrosoftAuthenticateConfig()
	address := utils.GetCurrentApDomain(username, microConfig)
	if address == "" {
		return errors.New("域不支持,请联系管理员")
	}
	ok, err := MicrosoftAuthenticate(address, username, password, microConfig.IsTLS)
	if err != nil {
		return fmt.Errorf("认证失败:%v", err)
	}
	if !ok {
		return fmt.Errorf("认证失败")
	}
	//通过用户名查询用户数据
	err = basedboperat.Get(a, nil, "username = ?", username)
	if err != nil {
		return err
	}
	a.Sso = true
	a.NewPassword(password)
	if a.ID == 0 {
		a.Username = username
		a.Nickname = username
		a.Identitys = append(a.Identitys, identityObject.Identity{ID: 2})
		basedboperat.Create(a)
		a.CreateRelation()
	} else {
		basedboperat.Update(a, nil, "id = ?", a.ID)
		basedboperat.Update(a, []string{"sso"}, "id = ?", a.ID)
	}
	//生成用户信息
	a.UserToken = NewToken(a)

	return nil
}

// ldap://172.0.0.54:389 fanhan\\administrator Password1!
func MicrosoftAuthenticate(address, username, password string, isTls bool) (bool, error) {
	// 连接到 LDAP 服务器
	// 创建一个连接到 LDAP 服务器的 TLS 配置
	var l *ldap.Conn
	var err error
	if isTls {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // 注意：在生产环境中应谨慎使用，最好使用有效证书
		}
		l, err = ldap.DialTLS("tcp", address, tlsConfig)
		if err != nil {
			return false, err
		}
		defer l.Close()
	} else {
		l, err = ldap.Dial("tcp", address)
		if err != nil {
			return false, err
		}
		defer l.Close()
	}
	// 绑定使用完整的 DN 格式
	// bindDN := fmt.Sprintf("%s", username) // 使用 fanhan\administrator 格式
	err = l.Bind(username, password)
	if err != nil {
		fmt.Println(err)
		return false, nil // 返回 nil 代表凭证无效
	}
	return true, nil // 返回 true 代表认证成功
}

// 用户注销
func (a *Account) Logoff() {
	DeleteToken(a.UserToken.TokenKey)
}

func LoginByToken(tokenKey string) (*Account, error) {
	userToken, err := GetToken(tokenKey)
	if err != nil {
		return nil, err
	}
	sessionAccount := userToken.Payload.(*Account)
	sessionAccount.UserToken = userToken
	return sessionAccount, nil
}

func (a *Account) LoginByToken(tokenKey string) (*Account, error) {
	userToken, err := GetToken(tokenKey)
	if err != nil {
		return nil, err
	}
	sessionAccount := userToken.Payload.(*Account)
	sessionAccount.UserToken = userToken
	return sessionAccount, nil
}

func (a *Account) CheckToken(s string) bool {
	//判断token是否正确
	if s != a.UserToken.TokenKey {
		return false
	}
	//再判断token是否过期
	return a.UserToken.Valid()
}
func (a *Account) Get() error {

	if a.ID > 0 {
		return basedboperat.Get(a, nil, "id = ?", a.ID)
	}
	if a.Username != "" {
		return basedboperat.Get(a, nil, "username = ?", a.Username)
	}
	return errors.New("id或username不能为空")
}

// 修改密码
func (a *Account) Changepassword(oldpassword, newpassword string) error {

	//通过用户名查询用户数据

	err := basedboperat.Get(a, nil, "id = ?", a.ID)
	if err != nil {
		return err
	}
	//检查用户名是否存在

	//判断用户是否存在
	if a.Username == "" {
		return errors.New("用户名不存在")
	}

	//加密密码
	upassword := system.SHA1(oldpassword)

	//撒盐
	upassword = upassword + a.Salt

	//混合加密
	upassword = system.SHA1(upassword)

	//判断密码是否一致
	if a.Password != upassword {
		//不一致:返回错误
		return errors.New("原始密码错误")
	}

	a.NewPassword(newpassword)

	basedboperat.Update(a, []string{"password", "salt"}, "id = ?", a.ID)

	return nil
}

func (a *Account) Update() error {
	var findModel Account
	basedboperat.Get(&findModel, nil, "id = ?", a.ID)
	if findModel.ID <= 0 {
		return errors.New("记录不存在")
	}
	a.DeleteRelation()
	_, err := a.CreateRelation()
	if err != nil {
		return err
	}
	basedboperat.Update(a, []string{"nickname", "email", "weixin_userid", "mobile", "we_com_id"}, "id = ?", a.ID)
	return nil
}

func (a *Account) Delete() (int64, error) {
	var findModel Account
	basedboperat.Get(&findModel, nil, "id = ?", a.ID)
	if !findModel.CanDelete {
		return 0, errors.New("无法删除系统用户")
	}
	a.DeleteRelation()
	return basedboperat.Delete(a, a.ID, nil)
}

// 删除关联数据
func (a *Account) DeleteRelation() {
	if a.ID == 0 {
		return
	}
	var ia identityObject.IdentityAccounts
	var ij DepartmentAccounts
	basedboperat.SqlExec("delete from "+ia.TableName()+" where account_id = ?", a.ID)
	basedboperat.SqlExec("delete from "+ij.TableName()+" where account_id = ?", a.ID)
}

// 创建关联数据
func (a *Account) CreateRelation() (int64, error) {
	if a.ID == 0 {
		return 0, nil
	}
	var rowsAffected int64
	for _, v := range a.Departments {
		var da DepartmentAccounts
		da.AccountId = a.ID
		da.DepartmentId = v.ID
		rowsAffected2, _ := basedboperat.Create(&da)
		rowsAffected = rowsAffected + rowsAffected2
	}

	for _, v := range a.Identitys {
		var id identityObject.Identity
		if v.ID == 0 {
			if v.Code == "" {
				continue
			}
			basedboperat.Get(&id, nil, "code = ?", v.Code)
		} else {
			basedboperat.Get(&id, nil, "id = ?", v.ID)
		}

		if id.ID == 0 {
			continue
		}

		var j identityObject.IdentityAccounts
		j.AccountId = a.ID
		j.IdentityId = id.ID
		rowsAffected2, _ := basedboperat.Create(&j)
		rowsAffected = rowsAffected + rowsAffected2
	}
	return rowsAffected, nil
}

func (a *Account) AfterQuery() {
	if a.ID == 0 {
		return
	}
	//中间表查询
	var ia identityObject.IdentityAccounts
	var ias []identityObject.IdentityAccounts
	var da DepartmentAccounts
	var das []DepartmentAccounts

	basedboperat.SqlQueryScan(&ias, "select * from "+ia.TableName()+" where account_id = ?", a.ID)
	basedboperat.SqlQueryScan(&das, "select * from "+da.TableName()+" where account_id = ?", a.ID)

	//关联部门
	var dp_list basedboperat.List
	dp_list.Limit = -1
	dp_list.And = map[string][]interface{}{}
	department_ids := []interface{}{}
	var department Department
	var departments []Department
	if len(das) > 0 {
		for _, v := range das {
			department_ids = append(department_ids, v.DepartmentId)
		}
		dp_list.And["id"] = department_ids
		basedboperat.ListScan(&dp_list, department, &departments)
	}

	a.Departments = departments

	//关联身份
	var identity_list basedboperat.List
	identity_list.Limit = -1
	identity_list.And = map[string][]interface{}{}
	identity_ids := []interface{}{}
	var identity identityObject.Identity
	var identitys []identityObject.Identity
	if len(ias) > 0 {
		for _, v := range ias {
			identity_ids = append(identity_ids, v.IdentityId)
		}
		identity_list.And["id"] = identity_ids
		basedboperat.ListScan(&identity_list, identity, &identitys)
	}

	a.Identitys = identitys

	//关联权限
	var ij identityObject.IdentityJurisdictions
	var ijs []identityObject.IdentityJurisdictions

	var identity_jurisdictions_list basedboperat.List
	identity_jurisdictions_list.Limit = -1
	identity_jurisdictions_list.And = map[string][]interface{}{}
	identity_jurisdictions_list_ids := []interface{}{}
	if len(identitys) > 0 {
		for _, v := range identitys {
			identity_jurisdictions_list_ids = append(identity_jurisdictions_list_ids, v.ID)
		}
		identity_jurisdictions_list.And["identity_id"] = identity_jurisdictions_list_ids
		basedboperat.ListScan(&identity_jurisdictions_list, ij, &ijs)
	}

	if len(ijs) > 0 {
		var jj jurisdictionObject.Jurisdiction
		var jjs []jurisdictionObject.Jurisdiction
		var jurisdictions_list basedboperat.List
		jurisdictions_list.Limit = -1
		jurisdictions_list.And = map[string][]interface{}{}
		jurisdictions_list_ids := []interface{}{}
		for _, v := range ijs {
			jurisdictions_list_ids = append(jurisdictions_list_ids, v.JurisdictionId)
		}
		jurisdictions_list.And["id"] = jurisdictions_list_ids
		basedboperat.ListScan(&jurisdictions_list, jj, &jjs)
		a.Jurisdictions = jjs
	}

}

// 注册外部账户
func (a *Account) RegAndUpdateOauth(username, nickname, mail, access_token, uid string) error {

	//数据库查询出用户信息

	findAccount := Account{}
	basedboperat.Get(&findAccount, nil, "username = ?", username)
	if findAccount.ID == 0 {
		return errors.New("当前用户没有权限登录,请联系管理员先创建用户")
	}
	if findAccount.GetAccountActivateStatus() == "disable" {
		return errors.New("当前用户已被禁用,请联系管理员先激活用户")
	}
	// //检查用户名是否存在
	// if findAccount.ID != 0 {
	// 	//更新auth数据
	findAccount.NewPassword(access_token)
	findAccount.Sso = true
	err := basedboperat.Update(&findAccount, []string{"password", "sso", "salt"}, "username = ?", username)
	if err != nil {
		return err
	}
	// 	return nil
	// }

	// a.Sso = true
	// a.Nickname = nickname
	// a.Email = mail
	// a.AccessToken = access_token

	// a.MsUid = uid

	// //生成密码
	// a.NewPassword(access_token)

	// a.Username = username
	// a.CanDelete = true
	// //创建权限

	// a.Identitys = append(a.Identitys, identityObject.Identity{ID: 2})

	// _, err := basedboperat.Create(&a) // 通过数据的指针来创建
	// if err != nil {
	// 	return err
	// }
	// _, err = a.CreateRelation()
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (a *Account) GetAccountActivateStatus() string {
	var accountActivate AccountActivateStatus
	basedboperat.Get(&accountActivate, nil, "account_id = ?", a.ID)

	return accountActivate.Status
}

func (a *Account) DeleteVerify() error {
	if a.ID == 0 {
		return errors.New("id不合法")
	}
	return nil
}

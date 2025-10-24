package accountcenter

import (
	"log"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/module/accountcenter/identityObject"
	"github.com/towgo/towgo/module/accountcenter/jurisdictionObject"
)

func initLoader() {
	//注册拦截器
	initInterceptor()

	//部门管理API
	initDepartmentApi()

	//SSO 模块
	initSSOApi()

	//身份（角色）模块
	identityObject.InitIdentityApi()

	//权限模块
	jurisdictionObject.InitjurisdictionApi()

	//token任务
	InitTokenTask()

	//导航栏
	InitNav()

	xormDriver.Sync2(new(Account), new(Department), new(DepartmentAccounts), new(DepartmentJurisdiction))
	xormDriver.Sync2(new(SSOToken))
	xormDriver.Sync2(new(UserToken))
	xormDriver.Sync2(new(identityObject.Identity), new(identityObject.IdentityAccounts), new(identityObject.IdentityJurisdictions), new(identityObject.IdentityNavs))
	xormDriver.Sync2(new(jurisdictionObject.Jurisdiction))
	xormDriver.Sync2(new(AccountActivateStatus), new(Group), new(AccountGroupRelation))

	autoCreateDefaultData()
}

func autoCreateDefaultData() {
	time.Sleep(time.Second * 5)
	//创建默认权限
	var j jurisdictionObject.Jurisdiction
	basedboperat.Get(&j, nil, "id = ?", -1)
	if j.ID != -1 {
		j.ID = -1
		j.Code = "superadmin"
		j.Name = "超级管理员"
		basedboperat.Create(&j)
	}

	//创建默认角色
	var identity identityObject.Identity
	basedboperat.Get(&identity, nil, "id = ?", 1)
	if identity.ID != 1 {
		identity.ID = 1
		identity.Code = "superadmin"
		identity.IsSystem = true
		identity.Name = "超级管理员"
		identity.Remark = "系统默认"
		basedboperat.Create(&identity)
	}

	//创建默认用户
	var account Account
	basedboperat.Get(&account, nil, "id = ?", 1)
	if account.ID != 1 {
		account.ID = 1
		account.Nickname = "管理员"
		account.Username = "admin"
		account.CanDelete = false
		account.NewPassword("123456")
		account.Identitys = append(account.Identitys, identity)
		_, err := basedboperat.Create(&account) // 通过数据的指针来创建
		if err != nil {
			log.Print(err.Error())
		}

		var ia identityObject.IdentityAccounts
		ia.AccountId = 1
		ia.ID = 1
		ia.IdentityId = 1
		basedboperat.Create(&ia)

		var ij identityObject.IdentityJurisdictions
		ij.ID = 1
		ij.IdentityId = 1
		ij.JurisdictionId = -1
		basedboperat.Create(&ij)
	}

}

package accountcenter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/dblog"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/sliceOperate"
	"github.com/towgo/towgo/module/accountcenter/accountctx"
	"github.com/towgo/towgo/module/accountcenter/identityObject"
	// "github.com/towgo/towgo/module/dblog"
)

func InitManageApi() {

	//初始化API加载器
	initLoader()

	//注册JSON-RPC服务处理器method路由
	//账户登录 F
	jsonrpc.SetFunc("/account/login", accountLogin)

	//获取自己的账户信息
	jsonrpc.SetFunc("/account/myinfo", accountMyinfo)
	jsonrpc.SetFunc("/account/myImg", accountMyImg)
	jsonrpc.SetFunc("/account/updateMyImg", accountUpdateMyImg)
	//账户注销 F
	jsonrpc.SetFunc("/account/logoff", accountLogoff)
	//账户注册 F
	jsonrpc.SetFunc("/account/reg", accountReg)
	//账户详情查询 F
	jsonrpc.SetFunc("/account/query", accountQuery)
	//账户列表查询
	jsonrpc.SetFunc("/account/list", accountList)

	jsonrpc.SetFunc("/account/listbyidentity", accountListByIdentity)

	//修改账户信息 F
	jsonrpc.SetFunc("/account/update", accountUpdate)
	jsonrpc.SetFunc("/account/updateMyinfo", updateMyinfo)
	jsonrpc.SetFunc("/account/updateNickname", updateNickname)

	jsonrpc.SetFunc("/account/resetpassword", jsonrpc_api_reset_password)
	//删除账户 F
	jsonrpc.SetFunc("/account/delete", accountDelete)

	//修改密码 F
	jsonrpc.SetFunc("/account/changepassword", accountChangepassword)

	//token验证 F
	jsonrpc.SetFunc("/account/token/check", accountTokenCheck)

	jsonrpc.SetFunc("/account/islogin", islogin)

	jsonrpc.SetFunc("/account/identity/query", identityQuery)

	jsonrpc.SetFunc("/account/UnauthorizedMethodList", accountUnauthorizedMethodList)

	jsonrpc.SetFunc("/account/UnauthorizedMethodAdd", accountUnauthorizedMethodAdd)

	jsonrpc.SetFunc("/account/UnauthorizedMethodDel", accountUnauthorizedMethodDel)

	// 通过用户名列表获取用户详细信息
	jsonrpc.SetFunc("/account/getAccountInfoByUsername", getAccountInfoByUsername)
	// 通过部门id获取下面的所有用户包括子集部门
	jsonrpc.SetFunc("/get/account/under/department", getAccountUnderDepartment)
	jsonrpc.SetFunc("/get/account/under/departmentGroup", getAccountUnderDepartmentGroup)

	// 创建第三方大量用户
	jsonrpc.SetFunc("/account/third/create/orupdate", CreateOrUpdateThirdPlentyUser)
	// 激活和禁用用户
	jsonrpc.SetFunc("/account/activate", accountActivate)

	jsonrpc.SetFunc("/group/create", groupCreate)
	jsonrpc.SetFunc("/group/update", groupUpdate)
	jsonrpc.SetFunc("/group/list", groupList)
	jsonrpc.SetFunc("/group/delete", groupDelete)

	jsonrpc.SetFunc("/account/group/relate", accountGroupRelate)
	jsonrpc.SetFunc("/account/group/info", accountGroupInfo)

	UnauthorizedMethodAdd("/account/third/create/orupdate")

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.LOGIN, "/account/login", "account_center", "登录"),
		dblog.NewOperateType(dblog.QUERY, "/account/myinfo", "account_center", "查询个人账户信息"),
		dblog.NewOperateType(dblog.QUERY, "/account/myImg", "account_center", "查询个人账户头像"),
		dblog.NewOperateType(dblog.UPDATE, "/account/updateMyImg", "account_center", "修改头像"),
		dblog.NewOperateType(dblog.LOGIN_OUT, "/account/logoff", "account_center", "登出"),
		dblog.NewOperateType(dblog.CREATE, "/account/reg", "account_center", "创建账户"),
		dblog.NewOperateType(dblog.QUERY, "/account/query", "account_center", "查询账户详情"),
		dblog.NewOperateType(dblog.QUERY, "/account/list", "account_center", "查询账户列表"),
		dblog.NewOperateType(dblog.QUERY, "/account/listbyidentity", "account_center", "查询角色列表"),
		dblog.NewOperateType(dblog.UPDATE, "/account/update", "account_center", "修改账户信息"),
		dblog.NewOperateType(dblog.UPDATE, "/account/updateMyinfo", "account_center", "修改个人账户信息"),
		dblog.NewOperateType(dblog.UPDATE, "/account/updateNickname", "account_center", "修改个人昵称"),
		dblog.NewOperateType(dblog.UPDATE, "/account/resetpassword", "account_center", "重置密码"),
		dblog.NewOperateType(dblog.DELETE, "/account/delete", "account_center", "删除账户"),
		dblog.NewOperateType(dblog.UPDATE, "/account/changepassword", "account_center", "修改个人密码"),
		dblog.NewOperateType(dblog.QUERY, "/account/token/check", "account_center", "令牌检查"),
		dblog.NewOperateType(dblog.QUERY, "/account/islogin", "account_center", "查询登录状态"),
		dblog.NewOperateType(dblog.QUERY, "/account/identity/query", "account_center", "查询角色信息"),
		dblog.NewOperateType(dblog.QUERY, "/account/UnauthorizedMethodList", "account_center", "查询鉴权白名单"),
		dblog.NewOperateType(dblog.CREATE, "/account/UnauthorizedMethodAdd", "account_center", "增加鉴权白名单"),
		dblog.NewOperateType(dblog.DELETE, "/account/UnauthorizedMethodDel", "account_center", "删除鉴权白名单"),
		dblog.NewOperateType(dblog.QUERY, "/account/getAccountInfoByUsername", "account_center", "根据账户名查询个人信息"),
		dblog.NewOperateType(dblog.QUERY, "/get/account/under/department", "account_center", "查询部门信息"),
		dblog.NewOperateType(dblog.QUERY, "/get/account/under/departmentGroup", "account_center", "查询部门组"),
		dblog.NewOperateType(dblog.UPDATE, "/account/third/create/orupdate", "account_center", ""),
		dblog.NewOperateType(dblog.UPDATE, "/account/activate", "account_center", ""),
		dblog.NewOperateType(dblog.CREATE, "/group/create", "account_center", "创建组"),
		dblog.NewOperateType(dblog.UPDATE, "/group/update", "account_center", "修改组"),
		dblog.NewOperateType(dblog.QUERY, "/group/list", "account_center", "查询组列表"),
		dblog.NewOperateType(dblog.DELETE, "/group/delete", "account_center", "删除组"),
		dblog.NewOperateType(dblog.UPDATE, "/account/group/relate", "account_center", "创建组关联"),
		dblog.NewOperateType(dblog.QUERY, "/account/group/info", "account_center", "查看组信息"),
	)

}

// 根据角色查询用户列表
func accountListByIdentity(rpcConn jsonrpc.JsonRpcConnection) {
	var params identityObject.Identity
	var list basedboperat.List
	rpcConn.ReadParams(&params, &list)

	if params.ID != 0 {
		var account Account
		var accounts []Account
		var id identityObject.Identity
		basedboperat.Get(&id, nil, "id = ?", params.ID)
		if id.ID > 0 {
			if list.And == nil {
				list.And = map[string][]interface{}{}
				for _, v := range id.Accounts {
					list.And["id"] = append(list.And["id"], v)
				}
			} else {
				for _, v := range id.Accounts {
					list.And["id"] = append(list.And["id"], v)
				}
			}
			basedboperat.ListScan(&list, &account, &accounts)
		}

		result := map[string]interface{}{}
		result["count"] = list.Count
		result["rows"] = accounts
		rpcConn.WriteResult(result)
		return
	}
	if params.Code != "" {
		var account Account
		var accounts []Account
		var id identityObject.Identity
		basedboperat.Get(&id, nil, "code = ?", params.Code)
		if id.ID > 0 {
			if list.And == nil {
				list.And = map[string][]interface{}{}
				for _, v := range id.Accounts {
					list.And["id"] = append(list.And["id"], v)
				}
			} else {
				for _, v := range id.Accounts {
					list.And["id"] = append(list.And["id"], v)
				}
			}
			basedboperat.ListScan(&list, &account, &accounts)
		}

		result := map[string]interface{}{}
		result["count"] = list.Count
		result["rows"] = accounts
		rpcConn.WriteResult(result)
		return
	}
	rpcConn.WriteError(500, "参数传递错误,如参必须有ID或CODE")
}

func accountUnauthorizedMethodList(rpcConn jsonrpc.JsonRpcConnection) {
	var list []string
	for k, _ := range unauthorizedMethod {
		list = append(list, k)
	}
	rpcConn.WriteResult(list)
}

func accountUnauthorizedMethodAdd(rpcConn jsonrpc.JsonRpcConnection) {
	var params struct {
		Methods []string
	}
	rpcConn.ReadParams(&params)
	for _, v := range params.Methods {
		UnauthorizedMethodAdd(v)
	}
	rpcConn.WriteResult("ok")
}

func accountUnauthorizedMethodDel(rpcConn jsonrpc.JsonRpcConnection) {
	var params struct {
		Methods []string
	}
	rpcConn.ReadParams(&params)
	for _, v := range params.Methods {
		UnauthorizedMethodDel(v)
	}
	rpcConn.WriteResult("ok")
}

func identityQuery(rpcConn jsonrpc.JsonRpcConnection) {

	var params struct {
		Identity string `json:"identity"`
	}
	rpcConn.ReadParams(&params)

	var account *Account
	account, err := account.LoginByToken(rpcConn.GetRpcRequest().Session)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}

	rpcConn.WriteResult(account.QueryIdentity(params.Identity))

}

func islogin(rpcConn jsonrpc.JsonRpcConnection) {

	var account *Account
	account, err := account.LoginByToken(rpcConn.GetRpcRequest().Session)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}

	if account.ID > 0 {
		rpcConn.WriteResult(true)
	} else {
		rpcConn.WriteResult(false)
	}

}

// 注册用户
func accountReg(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcresponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Account `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcresponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	if jsonObj.Params.Username == "" {
		rpcresponse.Error.Set(1001, "")
		rpcConn.Write()
		return
	}
	if jsonObj.Params.Password == "" {
		rpcresponse.Error.Set(1002, "")
		rpcConn.Write()
		return
	}

	/*
		account := Account{}
		account.Nickname = jsonObj.Params.Nickname
		account.Email = jsonObj.Params.Email
	*/
	if jsonObj.Params.Mobile != "" {
		if jsonObj.Params.Mobile != "" {
			getUserIdByPhoneResponse, err := GetQiYeWeiXinUserid(rpcConn.GetRpcRequest(), jsonObj.Params.Mobile)
			if err != nil {
				log.Printf("获取企业微信userid报错%s", err.Error())
			}
			if getUserIdByPhoneResponse.Errcode != 0 {
				log.Printf("获取企业微信userid报错%s", getUserIdByPhoneResponse.Errmsg)
			}
			jsonObj.Params.WeixinUserid = getUserIdByPhoneResponse.Userid
		}
	}
	Err := jsonObj.Params.Reg(jsonObj.Params.Username, jsonObj.Params.Password)
	if Err != nil {
		rpcresponse.Error.Set(1, Err.Error())
		rpcConn.Write()
		return
	}

	//拼装结果返回
	result["id"] = jsonObj.Params.ID
	result["username"] = jsonObj.Params.Username
	// dblog.Write("account:create","", fmt.Sprintf("用户名:%s,IP:%s",jsonObj.Params.Username,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult(result)
}

// 用户登录
func accountLogin(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}
	if jsonObj.Params.Username == "" {
		rpcResponse.Error.Set(1001, "")
		rpcConn.Write()
		return
	}
	if jsonObj.Params.Password == "" {
		rpcResponse.Error.Set(1002, "")
		rpcConn.Write()
		return
	}

	account := Account{}
	loginErr := account.Login(jsonObj.Params.Username, jsonObj.Params.Password)

	if loginErr != nil { //模型层登录成功
		// dblog.Write("account:login","", fmt.Sprintf("%s@%s 登录失败！ 错误信息:%s", account.Username, rpcConn.GetRemoteAddr(), loginErr.Error()))
		rpcResponse.Error.Set(1, "用户名或密码错误")
		rpcConn.Write()
		return
	}
	result["id"] = account.ID
	result["username"] = account.Username
	result["jurisdictions"] = account.Jurisdictions
	result["departments"] = account.Departments
	result["identitys"] = account.Identitys
	result["token"] = account.UserToken.TokenKey
	// dblog.Write("account:login","", fmt.Sprintf("%s@%s 登录成功！", account.Username, rpcConn.GetRemoteAddr()))
	go func() {
		for _, v := range account.Jurisdictions {
			if v.Code == "superadmin" {
				now := time.Now()
				//	发邮件
				request := rpcConn.GetRpcRequest()

				msg := fmt.Sprintf("数字员工平台访问地址 : %s </br> 超级管理员 : %s 登录系统  </br> IP : %s </br>  时间: %s", request.Route, account.Nickname, rpcConn.GetRemoteAddr(), now.Format("2006-01-02 15:04:05"))
				var temp struct {
					Email   []string `json:"email"`
					Message string   `json:"message"`
					Title   string   `json:"title"`
				}
				basePath := system.GetPathOfProgram()
				var mailS struct {
					AuditNotificationEmail []string `json:"AuditNotificationEmail"`
				}
				system.ScanConfigJson(filepath.Join(basePath, "config/mail.json"), &mailS)
				temp.Email = append(temp.Email, mailS.AuditNotificationEmail...)
				temp.Message = msg
				temp.Title = "数字员工平台超级管理员上线提醒"
				if len(temp.Email) == 0 {
					return
				}
				err1 := jsonrpc.Call("/account/sendEmailMassage", "", temp, nil)
				if err1 != nil {
					log.Println("邮件发送错误", err1.Error())
					return
				}
				break
			}
		}
	}()
	rpcConn.WriteResult(result)

}

// 列表查询接口
func accountList(rpcConn jsonrpc.JsonRpcConnection) {

	result := map[string]interface{}{
		"count": 0,
		"rows":  []interface{}{},
	}

	var list basedboperat.List

	rpcConn.ReadParams(&list)

	var account Account
	var accounts []Account

	departmentCondition := list.And["department"]
	if departmentCondition != nil {
		departmentAccountList := basedboperat.List{}
		var departmentAccount DepartmentAccounts
		var departmentAccounts []DepartmentAccounts
		departmentAccountList.And = map[string][]interface{}{}
		departmentAccountList.And["department_id"] = departmentCondition
		basedboperat.ListScan(&departmentAccountList, departmentAccount, &departmentAccounts)

		account_ids := []interface{}{}
		for _, v := range departmentAccounts {
			account_ids = append(account_ids, v.AccountId)
		}
		if len(account_ids) == 0 {
			account_ids = append(account_ids, 0)
		}
		list.And["id"] = account_ids
		delete(list.And, "department")
	}

	list.Field = []string{"id,username,email,nickname,created_at,updated_at,mobile,we_com_id"}

	basedboperat.ListScan(&list, account, &accounts)

	result["count"] = list.Count
	for i, v := range accounts {
		var accountActivateStatus AccountActivateStatus
		basedboperat.Get(&accountActivateStatus, nil, "account_id = ?", v.ID)
		if accountActivateStatus.Status == "disable" {
			accounts[i].AccountActivateStatus = "disable"
		} else {
			accounts[i].AccountActivateStatus = "activate"
		}
		var groupList []Group
		basedboperat.SqlQueryScan(&groupList, "select g.* from account_group_relation agr left join `group` g on agr.group_id = g.id where agr.account_id = ?", v.ID)
		accounts[i].Groups = groupList
	}
	result["rows"] = accounts
	rpcConn.WriteResult(result)

}

// token check
func accountTokenCheck(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Session string `json:"session"`
		Params  struct {
			Username string `json:"username"`
			Userid   int    `json:"userid"`
			Token    string `json:"token"`
		} `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	var account *Account
	account, err = account.LoginByToken(jsonObj.Params.Token)
	if err != nil {
		result["valid"] = false
		rpcConn.WriteResult(result)
		return
	}
	if account.ID == 0 {
		result["valid"] = false
		rpcConn.WriteResult(result)
		return
	}

	if !account.UserToken.Valid() {
		result["valid"] = false
		rpcConn.WriteResult(result)
		return
	}

	result["valid"] = true
	rpcConn.WriteResult(result)
}

// 用户注销
func accountLogoff(rpcConn jsonrpc.JsonRpcConnection) {
	//result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Session string `json:"session"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	var account *Account
	account, err = account.LoginByToken(jsonObj.Session)
	if err != nil {
		rpcConn.WriteResult(map[string]string{"success": "ok"})
		return
	}
	if account.ID > 0 {
		account.Logoff()
	}
	// dblog.Write("account:login","", fmt.Sprintf("%s@%s 登出！", account.Username, rpcConn.GetRemoteAddr()))
	go func() {
		for _, v := range account.Jurisdictions {
			if v.Code == "superadmin" {
				now := time.Now()
				request := rpcConn.GetRpcRequest()
				msg := fmt.Sprintf("数字员工平台访问地址 : %s </br> 超级管理员 : %s 退出系统  </br> IP : %s </br>  时间: %s", request.Route, account.Nickname, rpcConn.GetRemoteAddr(), now.Format("2006-01-02 15:04:05"))

				//	发邮件
				var temp struct {
					Email   []string `json:"email"`
					Message string   `json:"message"`
					Title   string   `json:"title"`
				}
				basePath := system.GetPathOfProgram()
				var mailS struct {
					AuditNotificationEmail []string `json:"AuditNotificationEmail"`
				}
				system.ScanConfigJson(filepath.Join(basePath, "config/mail.json"), &mailS)
				temp.Email = append(temp.Email, mailS.AuditNotificationEmail...)
				temp.Message = msg
				temp.Title = "数字员工平台超级管理员下线提醒"
				if len(temp.Email) == 0 {
					return
				}
				err1 := jsonrpc.Call("/account/sendEmailMassage", "", temp, nil)
				if err1 != nil {
					log.Println("邮件发送错误", err1.Error())
					return
				}
				break
			}
		}
	}()

	rpcConn.WriteResult(map[string]string{"success": "ok"})
}

func accountMyinfo(rpcConn jsonrpc.JsonRpcConnection) {
	accountSession, err := LoginByToken(rpcConn.GetRpcRequest().Session)
	accountSession.Token = rpcConn.GetRpcRequest().Session
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	var accountActivate AccountActivateStatus
	basedboperat.Get(&accountActivate, nil, "id = ?", accountSession.ID)
	if accountActivate.Status != "disable" {
		accountSession.AccountActivateStatus = "activate"
	} else {
		accountSession.AccountActivateStatus = accountActivate.Status
	}
	var groupList []Group
	basedboperat.SqlQueryScan(&groupList, "select g.* from account_group_relation agr left join `group` g on agr.group_id = g.id where agr.account_id = ?", accountSession.ID)
	accountSession.Groups = groupList
	var departmentList []Department
	basedboperat.SqlQueryScan(&departmentList, "select d.* from departments_accounts da left join departments d on da.department_id=d.id where da.account_id = ?", accountSession.ID)
	accountSession.Departments = departmentList
	rpcConn.WriteResult(accountSession)
}
func accountMyImg(rpcConn jsonrpc.JsonRpcConnection) {
	accountSession, err := LoginByToken(rpcConn.GetRpcRequest().Session)

	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	data := make(map[string]string)
	data["account_img"] = accountSession.AccountImg
	rpcConn.WriteResult(data)
}
func accountUpdateMyImg(rpcConn jsonrpc.JsonRpcConnection) {
	var param map[string]interface{}
	err := rpcConn.ReadParams(&param)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	accountImg := param["account_img"]
	if accountImg == "" {
		rpcConn.GetRpcResponse().Error.Set(500, "图片路径缺失")
		rpcConn.Write()
		return
	}
	accountSession, err := LoginByToken(rpcConn.GetRpcRequest().Session)
	accountSession.Token = rpcConn.GetRpcRequest().Session
	accountSession.AccountImg = accountImg.(string)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}

	err = basedboperat.Update(accountSession, []string{"account_img"}, "id = ?", accountSession.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}
func updateMyinfo(rpcConn jsonrpc.JsonRpcConnection) {
	var account Account
	rpcConn.ReadParams(&account)
	accountSession, err := LoginByToken(rpcConn.GetRpcRequest().Session)
	accountSession.Token = rpcConn.GetRpcRequest().Session
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	accountSession.Nickname = account.Nickname
	accountSession.Email = account.Email
	err = basedboperat.Update(accountSession, []string{"nickname", "email"}, "id = ?", accountSession.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	// dblog.Write("account:update",accountSession.Username, fmt.Sprintf("用户名:%s",accountSession.Username))
	rpcConn.WriteResult("ok")
}
func updateNickname(rpcConn jsonrpc.JsonRpcConnection) {
	var account Account
	rpcConn.ReadParams(&account)
	accountSession, err := LoginByToken(rpcConn.GetRpcRequest().Session)
	accountSession.Token = rpcConn.GetRpcRequest().Session
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	if account.Nickname == "" {
		account.Nickname = accountSession.Nickname
	}
	accountSession.Nickname = account.Nickname
	accountSession.Email = account.Email
	err = basedboperat.Update(accountSession, []string{"nickname"}, "id = ?", accountSession.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	// dblog.Write("account:update",accountSession.Username, fmt.Sprintf("用户名:%s",accountSession.Username))
	rpcConn.WriteResult("ok")
}

// 查询指定账户信息
func accountQuery(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{}
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Session string `json:"session"`
		Params  struct {
			Username string `json:"username"`
			Id       int64  `json:"id"`
			Token    string `json:"token"`
		} `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	var account *Account
	account, err = account.LoginByToken(jsonObj.Session)
	if err != nil {
		rpcResponse.Error.Set(401, err.Error())
		rpcConn.Write()
		return
	}
	if account.ID == 0 {
		rpcResponse.Error.Set(401, "未登录，无法访问")
		rpcConn.Write()
		return
	}

	//根据token查询
	if jsonObj.Params.Token != "" {
		var accountSession *Account
		accountSession, err = accountSession.LoginByToken(jsonObj.Params.Token)
		if err != nil {
			rpcResponse.Error.Set(401, err.Error())
			rpcConn.Write()
			return
		}

		if accountSession.ID > 0 {
			var accountActivate AccountActivateStatus
			basedboperat.Get(&accountActivate, nil, "id = ?", accountSession.ID)
			result["activate"] = accountActivate.Status
			result["id"] = accountSession.ID
			result["username"] = accountSession.Username
			result["nickname"] = accountSession.Nickname
			result["email"] = accountSession.Email
			result["identitys"] = accountSession.Identitys
			result["jurisdictions"] = accountSession.Jurisdictions
			result["departments"] = accountSession.Departments
			result["updatedAt"] = accountSession.UpdatedAt
			result["createdAt"] = accountSession.CreatedAt
			result["mobile"] = accountSession.Mobile
			result["we_com_id"] = accountSession.WeComId
			rpcConn.WriteResult(result)
			return
		} else {
			rpcResponse.Error.Set(1, "token不存在")
			rpcConn.Write()
			return
		}
	}

	var queryAccount Account

	queryAccount.ID = jsonObj.Params.Id
	queryAccount.Username = jsonObj.Params.Username
	err = queryAccount.Get()
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}
	var accountActivate AccountActivateStatus
	basedboperat.Get(&accountActivate, nil, "id = ?", queryAccount.ID)
	result["activate"] = accountActivate.Status
	result["id"] = queryAccount.ID
	result["username"] = queryAccount.Username
	result["nickname"] = queryAccount.Nickname
	result["email"] = queryAccount.Email
	result["identitys"] = queryAccount.Identitys
	result["jurisdictions"] = queryAccount.Jurisdictions
	result["departments"] = queryAccount.Departments
	result["updatedAt"] = queryAccount.UpdatedAt
	result["createdAt"] = queryAccount.CreatedAt
	result["mobile"] = queryAccount.Mobile
	result["we_com_id"] = queryAccount.WeComId
	rpcConn.WriteResult(result)
	return
}

func jsonrpc_api_reset_password(rpcConn jsonrpc.JsonRpcConnection) {

	currentAccount, err := accountctx.Parse(rpcConn)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}

	if !currentAccount.HasJurisdiction("superadmin") {
		rpcConn.GetRpcResponse().Error.Set(403, "权限不足")
		rpcConn.Write()
		return
	}

	var account Account
	rpcConn.ReadParams(&account)
	if account.ID == 0 {
		rpcConn.GetRpcResponse().Error.Set(500, "用户ID不能为空")
		rpcConn.Write()
		return
	}
	if account.Password == "" {
		rpcConn.GetRpcResponse().Error.Set(500, "密码不能为空")
		rpcConn.Write()
		return
	}

	//权限检查

	account.NewPassword(account.Password)
	err = basedboperat.Update(&account, []string{"password", "salt"}, "id = ?", account.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	var getAccount Account
	basedboperat.Get(&getAccount, nil, "id = ?", account.ID)
	// dblog.Write("account:reset_password",currentAccount.Username, fmt.Sprintf("用户名:%s", getAccount.Username))
	rpcConn.WriteResult("ok")
}

// 修改账户信息
func accountUpdate(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Session string  `json:"session"`
		Params  Account `json:"params"`
	}{}
	// currentAccount, err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}
	if jsonObj.Params.Mobile != "" {
		if jsonObj.Params.Mobile != "" {
			getUserIdByPhoneResponse, err := GetQiYeWeiXinUserid(rpcConn.GetRpcRequest(), jsonObj.Params.Mobile)
			if err != nil {
				log.Printf("获取企业微信userid报错%s", err.Error())
			}
			if getUserIdByPhoneResponse.Errcode != 0 {
				log.Printf("获取企业微信userid报错%s", getUserIdByPhoneResponse.Errmsg)
			}
			jsonObj.Params.WeixinUserid = getUserIdByPhoneResponse.Userid
		}
	}
	//写入数据库
	err = jsonObj.Params.Update()
	//err = basedboperat.Update(account, []string{"password", "nickname", "email", "salt"}, "id = ?", account.ID)
	if err != nil {
		result["success"] = false
		result["message"] = err.Error()
		rpcResponse.Error.Set(1, err.Error())
	} else {
		result["success"] = true
	}
	var getAccount Account
	basedboperat.Get(&getAccount, nil, "id= ?", jsonObj.Params.ID)
	// dblog.Write("account:update",currentAccount.Username, fmt.Sprintf("用户名:%s,IP:%s",getAccount.Username,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult(result)
}

func accountChangepassword(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Session string `json:"session"`
		Params  struct {
			Oldpassword string `json:"oldpassword"`
			Newpassword string `json:"newpassword"`
		} `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {

		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	var account *Account
	account, err = account.LoginByToken(jsonObj.Session)
	if err != nil {
		rpcResponse.Error.Set(401, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	if account.ID == 0 {
		rpcResponse.Error.Set(1003, "")
		rpcConn.WriteResult(result)
		return
	}

	err = account.Changepassword(jsonObj.Params.Oldpassword, jsonObj.Params.Newpassword)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}
	// dblog.Write("account:reset_password",account.Username, fmt.Sprintf("用户名:%s",account.Username))
	rpcConn.WriteResult(struct {
		Success string `json:"success"`
	}{Success: "ok"})

}

// 删除账户
func accountDelete(rpcConn jsonrpc.JsonRpcConnection) {

	var account Account
	rpcConn.ReadParams(&account)
	err := account.DeleteVerify()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	var sessionAccount *Account
	sessionAccount, err = sessionAccount.LoginByToken(rpcConn.GetRpcRequest().Session)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(403, err.Error())
		rpcConn.Write()
		return
	}

	if account.ID == sessionAccount.ID {
		rpcConn.GetRpcResponse().Error.Set(500, "无法删除自己的账号")
		rpcConn.Write()
		return
	}
	account.Get()
	if !account.CanDelete {
		rpcConn.GetRpcResponse().Error.Set(500, "该账号为系统保留，无法删除")
		rpcConn.Write()
		return
	}
	_, err = account.Delete()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	// dblog.Write("account:delete",sessionAccount.Username, fmt.Sprintf("用户名:%s,IP:%s", account.Username, rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult("ok")
}

type UsernameList struct {
	UsernameList []string `json:"username_list"`
}

func getAccountInfoByUsername(rpcConn jsonrpc.JsonRpcConnection) {
	var result []Account
	var usernameList UsernameList
	rpcConn.ReadParams(&usernameList)
	if len(usernameList.UsernameList) == 0 {
		rpcConn.WriteResult(result)
		return
	}
	var usernameQuoteList []string
	for _, username := range usernameList.UsernameList {
		usernameQuoteList = append(usernameQuoteList, "'"+username+"'")
	}
	sql := fmt.Sprintf("username in (%s)", strings.Join(usernameQuoteList, ","))
	err := basedboperat.QueryScan(&result, nil, sql)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(result)
}

func getAccountUnderDepartment(rpcConn jsonrpc.JsonRpcConnection) {
	var accountList []Account
	var departmentIdList DepartmentIdList
	rpcConn.ReadParams(&departmentIdList)
	if len(departmentIdList.IdList) == 0 {
		rpcConn.WriteResult(accountList)
		return
	}
	var idChildList []int64
	err := GetDepartmentChildIdList(departmentIdList.IdList, &idChildList)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	idChildList = append(idChildList, departmentIdList.IdList...)
	if len(idChildList) == 0 {
		rpcConn.WriteResult(accountList)
		return
	}
	var departAccount []DepartmentAccounts
	idChildStrList := TranslateSqlInAny(idChildList)
	sql := fmt.Sprintf("select * from departments_accounts where department_id in (%s)", strings.Join(idChildStrList, ","))
	basedboperat.SqlQueryScan(&departAccount, sql)
	var accountIdList []int64
	for _, v := range departAccount {
		accountIdList = append(accountIdList, v.AccountId)
	}
	if len(accountIdList) == 0 {
		rpcConn.WriteResult(accountList)
		return
	}
	accountIdStrList := TranslateSqlInAny(accountIdList)
	sql = fmt.Sprintf("id in (%s)", strings.Join(accountIdStrList, ","))
	basedboperat.QueryScan(&accountList, nil, sql)
	rpcConn.WriteResult(accountList)
}

func getAccountUnderDepartmentGroup(rpcConn jsonrpc.JsonRpcConnection) {
	var departments []Department

	var departmentIdList DepartmentIdList
	rpcConn.ReadParams(&departmentIdList)
	if len(departmentIdList.IdList) == 0 {
		rpcConn.WriteResult(departments)
		return
	}
	for _, v := range departmentIdList.IdList {
		users, err := getDepartmentChildDeptAndUsers(v)
		if err != nil {
			rpcConn.WriteResult(err)
			return
		}
		departments = append(departments, users)
	}
	rpcConn.WriteResult(departments)
}

func getDepartmentChildDeptAndUsers(deptId int64) (Department, error) {
	//查询部门

	var department Department

	err := basedboperat.Get(&department, nil, "id = ?", deptId)
	if err != nil {
		log.Println("部门查询", err)
		return department, err
	}
	//	查询子部门

	var childDepartIdList []Department
	sql := fmt.Sprintf("select id from departments where fid = %d", deptId)
	err = basedboperat.SqlQueryScan(&childDepartIdList, sql)
	if err != nil {
		log.Println("查询子部门", err, sql)
		return department, err
	}

	//	查询用户
	var accountList []Account
	var departAccount []DepartmentAccounts
	sqlUsr := fmt.Sprintf("select * from departments_accounts where department_id = %d", deptId)
	err = basedboperat.SqlQueryScan(&departAccount, sqlUsr)
	if err != nil {
		log.Println("查询用户", err)
		return department, err
	}
	var accountIdList []int64
	for _, v := range departAccount {
		accountIdList = append(accountIdList, v.AccountId)
	}
	accountIdStrList := TranslateSqlInAny(accountIdList)
	sql = fmt.Sprintf("id in (%s)", strings.Join(accountIdStrList, ","))
	basedboperat.QueryScan(&accountList, nil, sql)
	department.Users = accountList
	var newChidlDeptList []Department
	for _, v := range childDepartIdList {
		users, err := getDepartmentChildDeptAndUsers(v.ID)
		if err != nil {
			log.Println("递归", err)
			return department, err
		}
		newChidlDeptList = append(newChidlDeptList, users)
		//v = users
	}
	department.ChildDepartment = newChidlDeptList
	return department, nil
}
func CreateOrUpdateThirdPlentyUser(rpcConn jsonrpc.JsonRpcConnection) {
	var needCreateorUpdateAccountList []Account
	rpcConn.ReadParams(&needCreateorUpdateAccountList)
	var getDepartmentList []Department
	basedboperat.SqlQueryScan(&getDepartmentList, "select * from departments")
	thirdDepartmentMap := make(map[string]Department)
	for _, v := range getDepartmentList {
		if v.ThirdId != "" {
			thirdDepartmentMap[v.ThirdId] = v
		}
	}
	var needBatchCreateAccountList []Account
	var needBatchCreateDepartAccountRelationList []DepartmentAccounts

	for idx, createorUpdateAccount := range needCreateorUpdateAccountList {
		if createorUpdateAccount.Username == "" {
			log.Println("需要添加的第三方用户名为空")
			continue
		}
		var itSelfUser Account
		basedboperat.Get(&itSelfUser, nil, "username = ?", createorUpdateAccount.Username)
		if itSelfUser.ID == 0 {
			needBatchCreateAccountList[idx].CreatedAt = time.Now().Unix()
			needBatchCreateAccountList = append(needBatchCreateAccountList, createorUpdateAccount)
		} else {
			if createorUpdateAccount.Nickname != itSelfUser.Nickname {
				basedboperat.Update(&createorUpdateAccount, []string{"nickname"}, "id = ?", itSelfUser.ID)
			}
		}
	}
	if len(needBatchCreateAccountList) > 0 {
		_, err := basedboperat.Create(&needBatchCreateAccountList)
		if err != nil {
			log.Println("创建第三方用户报错", err)
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	}
	// 统一创建部门和用户关联数据
	for _, createorUpdateAccount := range needCreateorUpdateAccountList {
		if createorUpdateAccount.Username == "" {
			log.Println("需要添加的第三方用户名为空")
			continue
		}
		var itSelfUser Account
		basedboperat.Get(&itSelfUser, nil, "username = ?", createorUpdateAccount.Username)
		if itSelfUser.ID == 0 {
			log.Printf("之前创建用户没有创建成功,用户名为%s\n", createorUpdateAccount.Username)
			continue
		}
		var getOldDepartAccount DepartmentAccounts
		basedboperat.Get(&getOldDepartAccount, nil, "account_id = ?", itSelfUser.ID)
		newDepart, ok := thirdDepartmentMap[createorUpdateAccount.ThirdDepartmentId]
		if !ok {
			log.Println("部门信息竟然未找到")
			continue
		}
		if getOldDepartAccount.ID == 0 {
			var needCreateDepartAccount DepartmentAccounts
			needCreateDepartAccount.DepartmentId = newDepart.ID
			needCreateDepartAccount.AccountId = itSelfUser.ID
			needBatchCreateDepartAccountRelationList = append(needBatchCreateDepartAccountRelationList, needCreateDepartAccount)
		} else {
			if getOldDepartAccount.DepartmentId != newDepart.ID {
				getOldDepartAccount.DepartmentId = newDepart.ID
				basedboperat.Update(&getOldDepartAccount, nil, "id = ?", getOldDepartAccount.ID)
				time.Sleep(1 * time.Second)
			}
		}
	}
	if len(needBatchCreateDepartAccountRelationList) > 0 {
		_, err := basedboperat.Create(&needBatchCreateDepartAccountRelationList)
		if err != nil {
			log.Println("创建第三方用户部门关联信息报错", err)
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	}
	rpcConn.WriteResult("ok")
}

func (a *AccountActivateStatus) AccountActivateVerify() error {
	if a.AccountId == 0 {
		return errors.New("account_id参数未上传")
	}
	if a.Status != "activate" && a.Status != "disable" {
		return errors.New("status不合法")
	}
	var account Account
	basedboperat.Get(&account, nil, "id = ?", a.AccountId)
	if account.ID == 0 {
		return errors.New("账号不存在")
	}
	return nil
}

// 激活和禁用用户
func accountActivate(rpcConn jsonrpc.JsonRpcConnection) {
	var accountActivateStatus AccountActivateStatus
	rpcConn.ReadParams(&accountActivateStatus)
	err := accountActivateStatus.AccountActivateVerify()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	var getAccountActivateStatus AccountActivateStatus
	basedboperat.Get(&getAccountActivateStatus, nil, "account_id = ?", accountActivateStatus.AccountId)
	if getAccountActivateStatus.ID == 0 {
		accountActivateStatus.ID = 0
		_, err = basedboperat.Create(&accountActivateStatus)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	} else {
		err = basedboperat.Update(&accountActivateStatus, nil, "id = ?", getAccountActivateStatus.ID)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	}
	var account Account
	basedboperat.Get(&account, nil, "id = ?", accountActivateStatus.AccountId)
	// dblog.Write("account:status",seesionAccount.Username, fmt.Sprintf("用户名:%s",account.Username))
	rpcConn.WriteResult("ok")
}

func groupCreate(rpcConn jsonrpc.JsonRpcConnection) {
	var group Group
	rpcConn.ReadParams(&group)
	group.ID = 0
	_, err := basedboperat.Create(&group)
	if err != nil {
		log.Println(err.Error())
		rpcConn.GetRpcResponse().Error.Set(500, "创建失败,请联系管理员")
		rpcConn.Write()
		return
	}
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("accountGroup:create",seesionAccount.Username, fmt.Sprintf("用户组名:%s,IP:%s",group.Name,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult("ok")
}

func groupUpdate(rpcConn jsonrpc.JsonRpcConnection) {
	var group Group
	rpcConn.ReadParams(&group)
	err := basedboperat.Update(&group, nil, "id = ?", group.ID)
	if err != nil {
		log.Println(err.Error())
		rpcConn.GetRpcResponse().Error.Set(500, "创建失败,请联系管理员")
		rpcConn.Write()
		return
	}
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("accountGroup:update",seesionAccount.Username, fmt.Sprintf("用户组ID:%d,用户组名:%s,IP:%s",group.ID,group.Name,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult("ok")
}

func groupDelete(rpcConn jsonrpc.JsonRpcConnection) {
	var group Group
	rpcConn.ReadParams(&group)
	_, err := basedboperat.Delete(&group, nil, "id = ?", group.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("accountGroup:delete",seesionAccount.Username, fmt.Sprintf("用户组ID:%d,IP:%s",group.ID,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult("ok")
}

func groupList(rpcConn jsonrpc.JsonRpcConnection) {
	var list basedboperat.List

	err := rpcConn.ReadParams(&list)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	var group Group
	var groupList []Group
	basedboperat.ListScan(&list, &group, &groupList)
	result := make(map[string]any)
	result["count"] = list.Count
	result["rows"] = groupList
	rpcConn.WriteResult(result)
}

type AccountGroupRelateParam struct {
	AccountId   int64   `json:"account_id"`
	GroupIdList []int64 `json:"group_id_list"`
}

func (a *AccountGroupRelateParam) accountGroupRelateVerify() error {
	if a.AccountId == 0 {
		return errors.New("用户未上传")
	}
	return nil
}

func accountGroupRelate(rpcConn jsonrpc.JsonRpcConnection) {
	var accountGroupRelateParam AccountGroupRelateParam
	rpcConn.ReadParams(&accountGroupRelateParam)
	err := accountGroupRelateParam.accountGroupRelateVerify()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	var groupList []Group
	basedboperat.SqlQueryScan(&groupList, "select g.* from account_group_relation agr left join `group` g on agr.group_id = g.id where agr.account_id = ?", accountGroupRelateParam.AccountId)
	var groupIdList []int64
	for _, v := range groupList {
		groupIdList = append(groupIdList, v.ID)
	}
	_, deleteList, createList := sliceOperate.GetListInterIntDiffer(groupIdList, accountGroupRelateParam.GroupIdList)
	if len(createList) != 0 {
		var groupCreateList []AccountGroupRelation
		for _, v := range createList {
			groupCreateList = append(groupCreateList, AccountGroupRelation{AccountId: accountGroupRelateParam.AccountId, GroupId: v})
		}
		_, err = basedboperat.Create(&groupCreateList)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	}
	if len(deleteList) != 0 {
		err = basedboperat.SqlExec(fmt.Sprintf("delete from account_group_relation where group_id in (%s) and account_id = ?", strings.Join(sliceOperate.TranslateSqlInAny(deleteList), ",")), accountGroupRelateParam.AccountId)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
	}
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// var account Account
	// basedboperat.Get(&account,nil,"id = ?",accountGroupRelateParam.AccountId)
	// dblog.Write("accountGroup:relate",seesionAccount.Username, fmt.Sprintf("用户名:%s,IP:%s",account.Username,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult("ok")
}

func accountGroupInfo(rpcConn jsonrpc.JsonRpcConnection) {
	var accountGroupRelateParam AccountGroupRelateParam
	rpcConn.ReadParams(&accountGroupRelateParam)
	err := accountGroupRelateParam.accountGroupRelateVerify()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	var groupList []Group
	basedboperat.SqlQueryScan(&groupList, "select g.* from account_group_relation agr left join `group` g on agr.group_id = g.id where agr.account_id = ?", accountGroupRelateParam.AccountId)
	rpcConn.WriteResult(groupList)
}

package accountcenter

import (
	"encoding/json"
	"github.com/towgo/towgo/module/dblog"

	// "fmt"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	// "github.com/towgo/towgo/module/accountcenter/accountctx"
	// "github.com/towgo/towgo/module/dblog"
)

func initDepartmentApi() {
	//组列表查询
	jsonrpc.SetFunc("/account/department/list", accountDepartmentList)

	//增加
	jsonrpc.SetFunc("/account/department/create", accountDepartmentCreate)

	//删除
	jsonrpc.SetFunc("/account/department/delete", accountDepartmentDelete)

	//修改
	jsonrpc.SetFunc("/account/department/update", accountDepartmentUpdate)

	//查询明细
	jsonrpc.SetFunc("/account/department/detail", accountDepartmentDetail)

	//树形结构
	jsonrpc.SetFunc("/account/department/treelist", accountDepartmentTreeList)

	//获取所有子孙级部门
	jsonrpc.SetFunc("/account/department/children", accountDepartmentChildList)
	// 第三方部门信息导入
	jsonrpc.SetFunc("/third/department/batch/create/orupdate", CreateOrUpdateThirdPlentyDepartment)

	UnauthorizedMethodAdd("/third/department/batch/create/orupdate")
	//获取下一级部门信息
	jsonrpc.SetFunc("/account/department/one/level/children/", getOneLevelDepartmentList)
	//获取当前部门下所有用户
	//jsonrpc.SetFunc("/account/department/getDepartmentUserList", getDepartmentUserList)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.QUERY, "/account/department/list", "account_center", "查询部门列表"),
		dblog.NewOperateType(dblog.CREATE, "/account/department/create", "account_center", "创建部门"),
		dblog.NewOperateType(dblog.DELETE, "/account/department/delete", "account_center", "删除部门"),
		dblog.NewOperateType(dblog.UPDATE, "/account/department/update", "account_center", "修改部门"),
		dblog.NewOperateType(dblog.QUERY, "/account/department/detail", "account_center", "查询部门详情"),
		dblog.NewOperateType(dblog.QUERY, "/account/department/treelist", "account_center", "查询部门树"),
		dblog.NewOperateType(dblog.QUERY, "/account/department/children", "account_center", "查询子部门"),
		dblog.NewOperateType(dblog.UPDATE, "/third/department/batch/create/orupdate", "account_center", "修改子部门"),
		dblog.NewOperateType(dblog.QUERY, "/account/department/one/level/children/", "account_center", "查看下一级子部门"),
	)
}
func accountDepartmentCreate(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Department `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	jsonObj.Params.ID = 0
	_, err = jsonObj.Params.Create()
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	result["id"] = jsonObj.Params.ID
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("department:create",seesionAccount.Username, fmt.Sprintf("部门名:%s,IP:%s",jsonObj.Params.Name,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult(result)
}

// 删除部门
func accountDepartmentDelete(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Department `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}
	var depart Department
	basedboperat.Get(&depart, nil, "id = ?", jsonObj.Params.ID)
	_, err = jsonObj.Params.Delete()
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	result["success"] = true
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("department:delete",seesionAccount.Username, fmt.Sprintf("部门ID:%d,部门名:%s,IP:%s",jsonObj.Params.ID,depart.Name,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult(result)
}

func accountDepartmentDetail(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Department `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	var department Department

	err = basedboperat.Get(&department, nil, "id = ?", jsonObj.Params.ID)
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	rpcConn.WriteResult(department)
}

// 修改
func accountDepartmentUpdate(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Department `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	err = jsonObj.Params.Update()
	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.WriteResult(result)
		return
	}

	result["success"] = true
	var depart Department
	basedboperat.Get(&depart, nil, "id = ?", jsonObj.Params.ID)
	// seesionAccount,err := accountctx.Parse(rpcConn)
	// if err != nil {
	// 	rpcConn.GetRpcResponse().Error.Set(500, err.Error())
	// 	rpcConn.Write()
	// 	return
	// }
	// dblog.Write("department:update",seesionAccount.Username, fmt.Sprintf("部门ID:%d,部门名:%s,IP:%s",jsonObj.Params.ID,jsonObj.Params.Name,rpcConn.GetRemoteAddr()))
	rpcConn.WriteResult(result)
}

// 列表查询接口
func accountDepartmentList(rpcConn jsonrpc.JsonRpcConnection) {

	result := map[string]interface{}{
		"count": 0,
		"rows":  []interface{}{},
	}

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		List basedboperat.List `json:"params"`
	}{}
	json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	var modelObject Department
	var modelObjects []Department

	basedboperat.ListScan(&jsonObj.List, modelObject, &modelObjects)

	if jsonObj.List.Error != nil {
		rpcResponse.Error.Set(2000, jsonObj.List.Error.Error())
		rpcConn.Write()
		return
	}

	result["count"] = jsonObj.List.Count
	result["rows"] = modelObjects

	rpcConn.WriteResult(result)

}

func accountDepartmentTreeList(rpcConn jsonrpc.JsonRpcConnection) {

	result := GetDepartmentTreeList()

	rpcConn.WriteResult(result)
}

type Params struct {
	DepartmentIds []int64 `json:"departmentIds"`
}

func accountDepartmentChildList(rpcConn jsonrpc.JsonRpcConnection) {
	department := struct {
		Params Params `json:"params"`
	}{}
	json.Unmarshal([]byte(rpcConn.Read()), &department)
	result := GetDepartmentTreeListId(department.Params.DepartmentIds, nil)
	rpcConn.WriteResult(result)

}

func CreateOrUpdateThirdPlentyDepartment(rpcConn jsonrpc.JsonRpcConnection) {
	var departmentList []Department
	rpcConn.ReadParams(&departmentList)
	err := CreateOrUpdateThirdDepartmentService(departmentList)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func getOneLevelDepartmentList(rpcConn jsonrpc.JsonRpcConnection) {
	var department Department
	rpcConn.ReadParams(&department)
	err := department.GetOneLevelDepartmentListVerify()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	if department.ID != 0 {
		var getDepartmentList []Department
		basedboperat.SqlQueryScan(&getDepartmentList, "select * from departments where fid = ? order by id", department.ID)
		for i, v := range getDepartmentList {
			var getChildrenDepartment Department
			var count int64
			basedboperat.Count(&getChildrenDepartment, &count, "fid = ?", v.ID)
			if count > 0 {
				getDepartmentList[i].HaveChildren = true
			}
		}
		rpcConn.WriteResult(getDepartmentList)
		return
	}
	if department.Name != "" || department.Nickname != "" {
		var getDepartmentList []Department
		basedboperat.SqlQueryScan(&getDepartmentList, "select * from departments where name like ? or nickname like ? order by id", "%"+department.Name+"%", "%"+department.Name+"%")
		for i, v := range getDepartmentList {
			var getChildrenDepartment Department
			var count int64
			basedboperat.Count(&getChildrenDepartment, &count, "fid = ?", v.ID)
			if count > 0 {
				getDepartmentList[i].HaveChildren = true
			}
		}
		rpcConn.WriteResult(getDepartmentList)
		return
	}
	var getDepartmentList []Department
	err = basedboperat.SqlQueryScan(&getDepartmentList, "select * from departments where fid = ? order by id", 0)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	for i, v := range getDepartmentList {
		var getChildrenDepartment Department
		var count int64
		basedboperat.Count(&getChildrenDepartment, &count, "fid = ?", v.ID)
		if count > 0 {
			getDepartmentList[i].HaveChildren = true
		}
	}
	rpcConn.WriteResult(getDepartmentList)
}

/*
func getDepartmentUserList(rpcConn jsonrpc.JsonRpcConnection) {
	result := map[string]interface{}{} //初始化结果参数
	var err error

	rpcResponse := rpcConn.GetRpcResponse()
	jsonObj := struct {
		Params Department `json:"params"`
	}{}
	err = json.Unmarshal([]byte(rpcConn.Read()), &jsonObj)

	if err != nil {
		rpcResponse.Error.Set(1, err.Error())
		rpcConn.Write()
		return
	}

	deptId := jsonObj.Params.ID

}*/

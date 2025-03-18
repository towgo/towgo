package jurisdictionObject

import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/dblog"
	// "github.com/towgo/towgo/module/accountcenter/identityObject"
)

var method string = "/account/jurisdiction"

var defaultModel Jurisdiction
var defaultModels []Jurisdiction

func InitjurisdictionApi() {
	jsonrpc.SetFunc(method+"/create", create)
	jsonrpc.SetFunc(method+"/delete", delete)
	jsonrpc.SetFunc(method+"/update", update)
	jsonrpc.SetFunc(method+"/detail", detail)
	jsonrpc.SetFunc(method+"/list", list)
	jsonrpc.SetFunc(method+"/treelist", treeList)

	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.CREATE, method+"/create", "account_center", "创建权限"),
		dblog.NewOperateType(dblog.DELETE, method+"/delete", "account_center", "删除权限"),
		dblog.NewOperateType(dblog.UPDATE, method+"/update", "account_center", "修改权限"),
		dblog.NewOperateType(dblog.QUERY, method+"/detail", "account_center", "查询权限详情"),
		dblog.NewOperateType(dblog.QUERY, method+"/list", "account_center", "查询权限列表"),
		dblog.NewOperateType(dblog.QUERY, method+"/treelist", "account_center", "查询权限树"),
	)

}

func create(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	newTransaction, err := basedboperat.NewTransaction()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		rpcConn.Write()
		return
	}
	err = newTransaction.Begin()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		rpcConn.Write()
		return
	}
	_, err = newTransaction.Create(&model)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		rpcConn.Write()
		return
	}
	// 查询角色中是否含有当前权限的父级,如果有,为这个角色添加当前新建的权限
	if model.Fid != 0 {
		var jurisdiction Jurisdiction
		newTransaction.Get(&jurisdiction, nil, "id = ?", model.Fid)
		if jurisdiction.ID == 0 {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, "父级id对应数据不存在")
			rpcConn.Write()
			return
		}
		var identityJurisdictionsObjectList []IdentityJurisdictionsObject
		err := newTransaction.SqlQueryScan(&identityJurisdictionsObjectList, "select * from identitys_jurisdictions where jurisdiction_id = ?", model.Fid)
		if err != nil {
			newTransaction.Rollback()
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
		for idx := range identityJurisdictionsObjectList {
			identityJurisdictionsObjectList[idx].JurisdictionId = model.ID
			identityJurisdictionsObjectList[idx].ID = 0
		}
		if len(identityJurisdictionsObjectList) > 0 {
			_, err = newTransaction.Create(&identityJurisdictionsObjectList)
			if err != nil {
				newTransaction.Rollback()
				rpcConn.GetRpcResponse().Error.Set(500, err.Error())
				rpcConn.Write()
				return
			}
		}

	}
	err = newTransaction.Commit()
	if err != nil {
		newTransaction.Rollback()
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func delete(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	newTransaction, err := basedboperat.NewTransaction()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		rpcConn.Write()
		return
	}
	err = newTransaction.Begin()
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
		rpcConn.Write()
		return
	}
	var identityJurisdictionsObject IdentityJurisdictionsObject
	_, err = newTransaction.Delete(&identityJurisdictionsObject, nil, "jurisdiction_id = ?", model.ID)
	if err != nil {
		newTransaction.Rollback()
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	_, err = newTransaction.Delete(&model, model.ID, nil)
	if err != nil {
		newTransaction.Rollback()
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	newTransaction.Commit()
	rpcConn.WriteResult("ok")
}

func update(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)
	err := basedboperat.Update(model, nil, "id = ?", model.ID)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult("ok")
}

func detail(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	rpcConn.ReadParams(&model)

	if model.ID == 0 {
		rpcConn.GetRpcResponse().Error.Set(500, "id 不能为空")
		rpcConn.Write()
		return
	}

	err := basedboperat.Get(&model, nil, nil)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(500, err.Error())
		rpcConn.Write()
		return
	}
	rpcConn.WriteResult(model)
}

func list(rpcConn jsonrpc.JsonRpcConnection) {
	model := defaultModel
	models := defaultModels
	var list basedboperat.List
	rpcConn.ReadParams(&list)
	basedboperat.ListScan(&list, model, &models)

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

func treeList(rpcConn jsonrpc.JsonRpcConnection) {
	model := Jurisdictions{}
	models := []Jurisdictions{}
	var list basedboperat.List
	rpcConn.ReadParams(&list)
	list.Limit = -1
	list.And = map[string][]interface{}{
		"fid": []interface{}{0},
	}
	basedboperat.ListScan(&list, model, &models)

	for k, v := range models {
		models[k].Childs = TreeList(v)
	}

	result := map[string]interface{}{}
	result["count"] = list.Count
	result["rows"] = models
	rpcConn.WriteResult(result)
}

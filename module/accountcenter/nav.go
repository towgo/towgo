package accountcenter

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/api"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/accountcenter/accountctx"
	"github.com/towgo/towgo/module/accountcenter/identityObject"
)

func (Nav) TableName() string {
	return "nav"
}

func InitNav() {
	xormDriver.Sync2(new(Nav))
	api.NewCRUDJsonrpcAPI("/nav", Nav{}, []Nav{}).RegAPI()
	jsonrpc.SetFunc("/nav/tree", nav_list_tree)
}

type Nav struct {
	ID       int64  `json:"id" xorm:"'ID' notnull pk autoincr comment('主键ID')"`
	ParentID int64  `json:"parent_id" xorm:"'parent_id' comment('父节点ID')"`
	Od       int64  `json:"od" xorm:"'od'  comment('排序')"`
	Name     string `json:"name" xorm:"'name' varchar(255) comment('栏目名称')"`
	Target   string `json:"target" xorm:"'target' varchar(255) comment('目标')"`
	Method   string `json:"method" xorm:"'method' varchar(255) comment('方法')"`
	Path     string `json:"path" xorm:"'path' varchar(255) comment('路径')"`
	IconUrl  string `json:"icon_url" xorm:"'icon' varchar(255) comment('图标url')"`
	IsEnable bool   `json:"is_enable" xorm:"'is_enable'   comment('是否可用 0/N,1/Y')"`
	Describe string `json:"describe" xorm:"'describe' varchar(2000) comment('导航栏描述')"`
	Children []Nav  `json:"children" xorm:"-"`

	searchWithIdentityID int64 `xorm:"-"`
	treeCount            int64 `xorm:"-"`
}

// 基于身份id查询树形结构
func nav_list_tree(rpcConn jsonrpc.JsonRpcConnection) {

	account, err := accountctx.Parse(rpcConn)
	if err != nil {
		rpcConn.WriteError(500, err.Error())
		return
	}

	var params struct {
		IdentityID int64 `json:"identity_id"`
	}
	err = rpcConn.ReadParams(&params)
	if err != nil {
		rpcConn.WriteError(500, err.Error())
		return
	}

	if !account.HasIdentity(params.IdentityID) {
		rpcConn.WriteError(500, "没有权限")
		return
	}

	var nav Nav
	var navs []Nav

	sql := "select * from " + nav.TableName() + " where parent_id = 0 and is_enable = 1"

	nav.searchWithIdentityID = params.IdentityID

	//如果设定条件为
	if nav.searchWithIdentityID > 0 {
		var in identityObject.IdentityNavs
		var ins []identityObject.IdentityNavs
		var inList basedboperat.List
		inList.Limit = -1
		inList.And = map[string][]interface{}{}
		inList.And["identity_id"] = []interface{}{nav.searchWithIdentityID}

		basedboperat.ListScan(&inList, &in, &ins)

		if len(ins) > 0 {
			var andValues []interface{}

			for _, v := range ins {
				andValues = append(andValues, v.NavID)
			}
			sql = sql + " and id in (" + InterfaceSliceToString(andValues) + ")"
		}

	}
	sql = sql + " order by od asc"
	err = basedboperat.SqlQueryScan(&navs, sql)
	if err != nil {
		rpcConn.WriteError(500, err.Error())
		return
	}

	for k, v := range navs {
		v.searchWithIdentityID = params.IdentityID
		children, err := v.TreeList(v)
		if err != nil {
			rpcConn.WriteError(500, err.Error())
			return
		}
		navs[k].Children = children
	}
	result := map[string]interface{}{}
	result["count"] = len(navs)
	result["rows"] = navs
	rpcConn.WriteResult(result)

}

func (n *Nav) TreeList(node Nav) ([]Nav, error) {

	if n.treeCount > 1000 {
		return nil, errors.New("递归超过限定次数,可能由于循环引用父级导致")
	}

	var nav Nav
	var navs []Nav

	sql := "select * from " + nav.TableName() + " where is_enable = 1 and parent_id = " + strconv.FormatInt(n.ID, 10)

	//如果设定条件为
	if n.searchWithIdentityID > 0 {
		var in identityObject.IdentityNavs
		var ins []identityObject.IdentityNavs
		var inList basedboperat.List
		inList.Limit = -1
		inList.And = map[string][]interface{}{
			"identity_id": []interface{}{n.searchWithIdentityID},
		}
		basedboperat.ListScan(&inList, in, &ins)

		if len(ins) > 0 {
			var andValues []interface{}

			for _, v := range ins {
				andValues = append(andValues, v.NavID)
			}
			sql = sql + " and id in (" + InterfaceSliceToString(andValues) + ")"
		}
	}

	sql = sql + " order by od asc"
	err := basedboperat.SqlQueryScan(&navs, sql)
	if err != nil {
		return nil, err
	}

	for k, v := range navs {
		v.searchWithIdentityID = n.searchWithIdentityID
		children, err := v.TreeList(v)
		if err != nil {
			return nil, err
		}
		navs[k].Children = children
	}
	return navs, nil
}

func InterfaceSliceToString(slice []interface{}) string {
	var strSlice []string
	for _, v := range slice {
		strSlice = append(strSlice, fmt.Sprintf("%v", v))
	}
	return strings.Join(strSlice, ",")
}

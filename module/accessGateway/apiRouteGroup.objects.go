package accessGateway

/*
统一接入网关 同步算法
by:liangliangit
*/
import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/accountcenter"
)

const (
	HEALTHY_CHECK_TYPE_HTTP = "http"
	HEALTHY_CHECK_TYPE_TCP  = "tcp"
	AUTH_NULL               = "AUTH_NULL"
	AUTH_ALL                = "AUTH_ALL"
)

var gateway_auth_mode string
var gateWay *AccessGateWay

var unauthorizedMethods []UnauthorizedMethod
var unauthorizedMethods_sync_sum string
var unauthorizedMethodMap sync.Map

var tempApiRouteGroups []ApiRouteGroup
var tempApiRouteGroups_sync_sum string

func init() {
	basedboperat.Sync(new(ApiRouteGroup), new(ApiRouteWay))
}

func StartServer(config AccessGateWayConfig) *AccessGateWay {
	gateWay = NewAccessGateWay(config)
	jsonrpcRouteInit(gateWay) //初始化jsonrpc路由
	go gateWay.ListenAndServe()
	ReLoad()
	go autoSync()
	return gateWay
}

func AuthMode(auth_mode string) error {
	switch auth_mode {
	case AUTH_ALL:
		gateway_auth_mode = AUTH_ALL
		return nil
	case AUTH_NULL:
		gateway_auth_mode = AUTH_NULL
		return nil
	default:
		return errors.New("模式不存在")
	}
}

func jsonrpcRouteInit(gateWay *AccessGateWay) {
	jsonrpc.OnMethodNotFound = (func(rpcConn jsonrpc.JsonRpcConnection) {
		request := rpcConn.GetRpcRequest()

		if gateway_auth_mode == AUTH_NULL {
			gateWay.TogoGateWayServer.ProxyToEdgeServerNode(rpcConn)
			request.Await()
			return
		}

		if request.Session != "" {
			rpcRequest := jsonrpc.NewJsonrpcrequest()
			rpcRequest.Session = request.Session
			rpcRequest.Method = "/account/myinfo"
			gateWay.TogoGateWayServer.CallEdgeServerNode(rpcRequest, func(jrc jsonrpc.JsonRpcConnection) {

				responseErr := jrc.GetRpcResponse().Error
				if responseErr.Code != 200 {
					rpcConn.GetRpcResponse().Error.Set(responseErr.Code, responseErr.Message)
					rpcConn.Write()
					return
				}

				var account accountcenter.Account
				jrc.ReadResult(&account)
				if account.ID == 0 {
					rpcConn.GetRpcResponse().Error.Set(401, "token不存在")
					rpcConn.Write()
					return
				}
				ctx := rpcConn.GetRpcRequest().Ctx
				if ctx == nil {
					ctx = map[jsonrpc.ContextKey]interface{}{}
				}
				ctx["account"] = account
				rpcConn.GetRpcRequest().Ctx = ctx
				gateWay.TogoGateWayServer.ProxyToEdgeServerNode(rpcConn)
				request.Await()
			})
			rpcRequest.Await()
		} else {
			gateWay.TogoGateWayServer.ProxyToEdgeServerNode(rpcConn)
			request.Await()
		}
	})
}

func (ApiRouteGroup) TableName() string {
	return "api_route_group"
}

type ApiRouteGroup struct {
	ID                     int64  `json:"id"`
	Name                   string `json:"name"`
	Domain                 string `json:"domain"`
	Schema                 string `json:"schema"`
	APIPath                string `json:"api_path"`
	APIType                string `json:"api_type"`
	IsWebFrontClient       bool   `json:"is_web_front_client"`
	UseTargetHostToRequest bool   `json:"use_target_host_to_request"`
	MaxConnectLimit        int64  `json:"max_connect_limit"` //最大并发量

	//负载均衡算法
	LoadType string `json:"load_type"`

	//运行池
	RunningPool []*ApiRouteWay `json:"-" gorm:"-" xorm:"-"`

	APIRouteWay []*ApiRouteWay `json:"api_route_way" gorm:"-" xorm:"-"`
}

// 删除前先删除关联信息
func (arw *ApiRouteGroup) BeforDelete(dbSession basedboperat.DbTransactionSession) error {
	var model ApiRouteWay
	_, err := dbSession.Delete(&model, nil, "group_id = ?", arw.ID)
	if err != nil {
		dbSession.Rollback()
		return err
	}
	return nil
}

func (arw *ApiRouteGroup) AfterSave(dbSession basedboperat.DbTransactionSession) error {

	if arw.ID == 0 {
		dbSession.Rollback()
		return errors.New("ApiRouteGroup can not be null")
	}

	err := dbSession.SqlExec("delete from "+arw.TableName()+"_way where group_id = ?", arw.ID)
	if err != nil {
		dbSession.Rollback()
		return err
	}

	for _, v := range arw.APIRouteWay {
		v.GroupId = arw.ID
		v.Schema = arw.Schema
		v.Domain = arw.Domain
		if arw.APIPath == "" {
			arw.APIPath = "/"
		}
		v.APIPath = arw.APIPath
		v.APIType = arw.APIType
		v.ID = 0
		_, err = dbSession.Create(v)
		if err != nil {
			log.Print(err.Error())
			dbSession.Rollback()
			return err
		}
	}
	return nil
}

func (arw *ApiRouteGroup) AfterQuery(dbSession basedboperat.DbTransactionSession) error {
	if len(dbSession.GetCurrentSelectFields()) > 0 {
		if !dbSession.IsCurrentSelectedField("group_id") {
			return nil
		}
	}

	var model ApiRouteWay
	var models []*ApiRouteWay
	var list basedboperat.List
	list.Limit = -1
	list.And = map[string][]interface{}{"group_id": []interface{}{arw.ID}}
	basedboperat.ListScan(&list, model, &models)
	arw.APIRouteWay = models
	return nil
}

func (ApiRouteWay) TableName() string {
	return "api_route_group_way"
}

type ApiRouteWay struct {
	ID                 int64  `json:"id"`
	GroupId            int64  `json:"group_id"`
	ApiRouteName       string `json:"api_route_name" xorm:"-" gorm:"-"`
	Domain             string `json:"domain"`
	Name               string `json:"name"`
	Schema             string `json:"schema"`
	Host               string `json:"host"`
	APIPath            string `json:"api_path"`
	HealthyCheckPath   string `json:"healthy_check_path"`
	HealthyCheckType   string `json:"healthy_check_type"`
	HealthyCheckEnable bool   `json:"healthy_check_enable"`
	UseApiBilling      bool   `json:"use_api_billing"`
	APIType            string `json:"api_type"`

	//是否运行，如果设定为false 那么 节点即使在线也不会参与代理转发，一般用于节点维护
	Running bool `json:"-" gorm:"-" xorm:"-"`
	Enable  bool `json:"enable"`

	IsOnline bool `json:"is_online"`

	/*
	  去除代理的头部路由 例如 代理路由为 /account 如果为true 代理作为客户端去请求后端服务器时
	  请求路径 /account/login 会去掉/account 保留/login作为请求
	*/
	RemoveProxyPath bool `json:"remove_proxy_path"`
}

func (arw *ApiRouteWay) AfterQuery() {
	var model ApiRouteGroup
	basedboperat.Get(&model, []string{"name"}, "id = ?", arw.GroupId)
	arw.ApiRouteName = model.Name
}

func (arw *ApiRouteWay) HealthyCheck() {
	if !arw.HealthyCheckEnable { //不进行健康检查
		return
	}
	switch arw.HealthyCheckType {
	case HEALTHY_CHECK_TYPE_TCP:
		conn, err := net.Dial("tcp", resolveIP(arw.Host))
		if err != nil {
			if arw.IsOnline {
				err = basedboperat.SqlExec("update "+arw.TableName()+" set is_online = 0 where id = ?", arw.ID)
				if err != nil {
					log.Print(err.Error())
				}
			}
		} else {
			if !arw.IsOnline {
				err = basedboperat.SqlExec("update "+arw.TableName()+" set is_online = 1 where id = ?", arw.ID)
				if err != nil {
					log.Print(err.Error())
				}
			}
			conn.Close()
		}
	default:

		resp, err := http.Get(arw.Schema + "://" + resolveIP(arw.Host) + arw.HealthyCheckPath)
		if err != nil {
			if arw.IsOnline {
				basedboperat.SqlExec("update "+arw.TableName()+" set is_online = 0 where id = ?", arw.ID)
			}
			return
		} else {
			//如果状态码不为零  那么认为服务正常
			if resp.StatusCode != 0 {
				if !arw.IsOnline {
					basedboperat.SqlExec("update "+arw.TableName()+" set is_online = 1 where id = ?", arw.ID)
				}
			} else {
				if arw.IsOnline {
					basedboperat.SqlExec("update "+arw.TableName()+" set is_online = 0 where id = ?", arw.ID)
				}
			}
		}
	}

}

func (arw *ApiRouteWay) GetTargetHost() (targetHost string) {
	targetHost = arw.Schema + "://" + resolveIP(arw.Host)
	return
}

func (arw *ApiRouteWay) GetTargetHostUrl() (targetHost string) {
	targetHost = arw.Schema + "://" + arw.Host
	return
}

func ReLoad() {
	go reload_apiRouteGroup()
	go reload_unauthorized_method()
}

func reload_apiRouteGroup() {
	// 获取数据库
	var apiRouteGroup ApiRouteGroup
	var apiRouteGroups []ApiRouteGroup
	var apiRouteGroupList basedboperat.List
	apiRouteGroupList.Limit = -1
	basedboperat.ListScan(&apiRouteGroupList, apiRouteGroup, &apiRouteGroups)

	newSum := system.MD5Any(apiRouteGroups)
	if newSum == tempApiRouteGroups_sync_sum {
		//没有需要同步的信息
		return
	}
	tempApiRouteGroups_sync_sum = newSum

	//检查变动信息

	// 删除和修改项
	deleteApiRouteWay := getDeleteOrUpdateApiRouteWay(apiRouteGroups)

	deleteApiRouteGroup := getDeleteOrUpdateApiRouteGroup(apiRouteGroups)

	for _, v := range deleteApiRouteWay {

		switch v.APIType {
		case "restful":
			err := gateWay.RemoveRestfulProxy(v.Domain, v.APIPath, v.GetTargetHostUrl())
			if err != nil {
				log.Print(err.Error())
			}
		case "jsonrpc":
			gateWay.RemoveJsonrpcProxy(v.Domain, v.APIPath, v.GetTargetHostUrl())
		}
	}

	for _, v := range deleteApiRouteGroup {
		gateWay.RemoveGroupProxy(v.Domain, v.APIPath, v.APIType)
	}

	for _, group := range apiRouteGroups {
		gateWay.NewGroupProxy(group.MaxConnectLimit, group.Domain, group.APIPath, group.APIType, group.LoadType, group.IsWebFrontClient)
		for _, way := range group.APIRouteWay {
			if !way.Enable {
				continue
			}
			if !way.IsOnline {
				continue
			}
			switch way.APIType {
			case "restful":
				err := gateWay.NewRestfulProxy(group, way)
				if err != nil {
					log.Print(err)
				}
			case "jsonrpc":
				err := gateWay.NewJsonrpcProxy(group.Domain, group.APIPath, way)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}

	tempApiRouteGroups = apiRouteGroups

}

func reload_unauthorized_method() {
	var model UnauthorizedMethod
	var models []UnauthorizedMethod
	var list basedboperat.List
	list.Limit = -1
	basedboperat.ListScan(&list, model, &models)

	newSum := system.MD5Any(models)
	if newSum == unauthorizedMethods_sync_sum {
		//没有需要同步的信息
		return
	}
	unauthorizedMethods_sync_sum = newSum
	unauthorizedMethods = models

	unauthorizedMethodMap.Range(func(key, value any) bool {
		unauthorizedMethodMap.Delete(key.(string))
		return true
	})

	for _, v := range unauthorizedMethods {
		unauthorizedMethodMap.Store(v.Method+v.Type, true)
	}
}

func getDeleteOrUpdateApiRouteGroup(apiRouteGroups []ApiRouteGroup) []ApiRouteGroup {
	// 检查删除项和修改项
	var deleteApiRouteGroup []ApiRouteGroup

	for _, tempGroup := range tempApiRouteGroups {
		var needDelete bool = true
		for _, remoteGroup := range apiRouteGroups {
			if tempGroup.ID != remoteGroup.ID {
				continue
			}
			if tempGroup.Domain != remoteGroup.Domain {
				break
			}
			if tempGroup.Schema != remoteGroup.Schema {
				break
			}
			if tempGroup.APIPath != remoteGroup.APIPath {
				break
			}
			if tempGroup.APIType != remoteGroup.APIType {
				break
			}
			if tempGroup.LoadType != remoteGroup.LoadType {
				break
			}
			if tempGroup.MaxConnectLimit != remoteGroup.MaxConnectLimit {
				break
			}
			if tempGroup.IsWebFrontClient != remoteGroup.IsWebFrontClient {
				break
			}

			if tempGroup.UseTargetHostToRequest != remoteGroup.UseTargetHostToRequest {
				break
			}

			needDelete = false
			break

		}
		if needDelete {
			deleteApiRouteGroup = append(deleteApiRouteGroup, tempGroup)
		}
	}

	return deleteApiRouteGroup
}

func getDeleteOrUpdateApiRouteWay(apiRouteGroups []ApiRouteGroup) []*ApiRouteWay {
	// 检查删除项和修改项
	var deleteApiRouteWay []*ApiRouteWay

	var tempApiRouteWay []*ApiRouteWay
	var remoteApiRouteWay []*ApiRouteWay

	for _, tmp_v := range tempApiRouteGroups {
		tempApiRouteWay = append(tempApiRouteWay, tmp_v.APIRouteWay...)
	}

	for _, remote_v := range apiRouteGroups {
		remoteApiRouteWay = append(remoteApiRouteWay, remote_v.APIRouteWay...)
	}

	for _, v := range tempApiRouteWay {

		var needDelete bool = true
		for _, v1 := range remoteApiRouteWay {
			if v.ID != v1.ID {
				continue
			}

			/*
				if system.MD5Any(v) != system.MD5Any(v1) {
					break
				}
			*/

			if v.Domain != v1.Domain {
				break
			}
			if v.APIPath != v1.APIPath {
				break
			}
			if v.APIType != v1.APIType {
				break
			}
			if v.Host != v1.Host {
				break
			}
			if v.Enable != v1.Enable {
				break
			}
			if v.IsOnline != v1.IsOnline {
				break
			}
			if v.HealthyCheckPath != v1.HealthyCheckPath {
				break
			}

			needDelete = false
			break

		}
		if needDelete {
			deleteApiRouteWay = append(deleteApiRouteWay, v)
		}
	}
	return deleteApiRouteWay
}

func resolveIP(input string) string {
	parts := strings.Split(input, ":")
	address := parts[0]
	port := "80"

	if len(parts) > 1 {
		port = parts[1]
	}

	ip := net.ParseIP(address)
	if ip == nil {
		// 如果无法解析为IP，则尝试解析为域名
		ips, err := net.LookupIP(address)
		if err != nil {
			fmt.Println("Unable to resolve the domain name")
			return ""
		}
		ip = ips[0]
	}
	target := ip.String() + ":" + port
	//log.Print("解析IP为:", target)
	return target
}

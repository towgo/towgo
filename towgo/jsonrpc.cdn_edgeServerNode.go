package towgo

/*
边缘服务节点
by:liangliangit
*/
import (
	"errors"
	"log"
	"sync"

	"github.com/towgo/towgo/lib/system"
)

var remoteGateWayServers sync.Map

type EdgeServerNode struct {
	Guid                 string
	OnConnect            func(*EdgeServerNode)
	OnClose              func(*EdgeServerNode)
	EdgeServerNodeConfig EdgeServerNodeConfig
	websocketConn        *WebScoketClient
}

type EdgeServerNodeConfig struct {
	Priority           int64    `json:"priority"`                        //优先级
	ModuleName         string   `json:"module_name"`                     //模块名称
	WebFrontServerPort string   `json:"web_front_server_port"`           //模块web服务端口 (前端web服务)
	Methods            []string `json:"methods" gorm:"json"`             //模块可提供的服务method
	DisableHealthCheck bool     `xorm:"-" gorm:"-"`                      //关闭心跳检测（默认服务端将启用心跳检测，在规定的时间内超时后服务端会主动端看）
	ServerUrl          string   `xorm:"-" gorm:"-"`                      //模块服务端url
	ClusterToken       string   `json:"cluster_token" xorm:"-" gorm:"-"` //集群token  如果token不正确不允许加入
	Schema             string   `json:"schema"`                          // eg: http | https
	Port               string   `json:"port"`
	Type               string   `json:"type"`
	EdgeServerNodeHost string   `json:"edge_server_node_host"`
	ServerUrls         []string `xorm:"-" gorm:"-"`
}

func NewEdgeServerNode(nodeConfig EdgeServerNodeConfig) *EdgeServerNode {
	serverNode := &EdgeServerNode{EdgeServerNodeConfig: nodeConfig, Guid: system.GetGUID().Hex()}
	return serverNode
}

func CallGateWay(method, token string, requestParams any, responseParams any) (err error) {
	err = errors.New("RPC调用失败,远程网关无法连接")
	request := NewJsonrpcrequest()
	request.Method = method
	request.Params = requestParams
	request.Session = token

	remoteGateWayServers.Range(func(key, value any) bool {
		edgeServerNode := value.(*EdgeServerNode)
		if edgeServerNode == nil {
			return false
		}

		edgeServerNode.Call(request, func(jrc JsonRpcConnection) {
			resp := jrc.GetRpcResponse()
			if resp.Error.Code != 200 {
				err = errors.New(resp.Error.Message)
				return
			}
			err = jrc.ReadResult(responseParams)
		})
		request.Await()
		return false
	})
	return
}

func (c *EdgeServerNode) Connect() {
	c.websocketConn = NewWebsocketClient(c.EdgeServerNodeConfig.ServerUrl, c.EdgeServerNodeConfig.ServerUrl)
	c.websocketConn.OnConnect = func(wsc *WebScoketClient) {
		remoteGateWayServers.Store(c.Guid, c)
		log.Print("网关连接成功 -> " + c.EdgeServerNodeConfig.ServerUrl)
		c.regModule()
		if c.OnConnect != nil {
			c.OnConnect(c)
		}

	}
	c.websocketConn.OnClose = func(wsc *WebScoketClient) {
		remoteGateWayServers.Delete(c.Guid)
		log.Print("网关断开链接 <- " + c.EdgeServerNodeConfig.ServerUrl)
		if c.OnClose != nil {
			c.OnClose(c)
		}
	}
	c.websocketConn.Connect()
}

func (c *EdgeServerNode) Call(request *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	c.websocketConn.Call(request, callback)
}

func (c *EdgeServerNode) regModule() {
	request := NewJsonrpcrequest()
	request.Method = "/" + API_HEAD + "/edgeServerNode/reg"
	request.Params = c.EdgeServerNodeConfig
	c.websocketConn.Push(request)
}

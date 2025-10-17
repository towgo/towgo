package towgo

/*
主控服务（网关）
by:liangliangit
*/
import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/www"
)

var API_HEAD = "towgocdn"

type GatewayServer struct {
	clusterToken                       string
	edgeServerNodeHttpPingTimeoutLimit int64
	sync.Mutex
	FrontServer                       www.WebServer
	loadAlgorithm                     func(string) JsonRpcConnection
	loadbalance                       LoadBalance
	WebsocketServer                   *WebSocketServer
	edgeServerNodeWebsocketMap        sync.Map
	edgeServerNodeHttpKeepalivedTimer sync.Map //map[string]*time.Timer
}

type GatewayEdgeServerNode struct {
	Guid                 string `json:"guid"  xorm:"index unique"`
	RemoteAddr           string `json:"remote_addr"`
	EdgeServerNodeConfig `xorm:"extends"`
}

func SetAPIHead(head string) {
	API_HEAD = head
}

func InitServerApi() {
	SetFunc("/"+API_HEAD+"/edgeServerNode/reg", reg)
	SetFunc("/"+API_HEAD+"/getEdgeServerNodeInfo", getEdgeServerNodeInfo)
	SetFunc("/"+API_HEAD+"/edgeServerNode/ping", ping)
	SetFunc("/"+API_HEAD+"/edgeServerNode/method/jsonrpcroute/list", edgeServerNodeMethodJsonrpcRouteList)
}

var gateWayServers []*GatewayServer

func edgeServerNodeMethodJsonrpcRouteList(rpcConn JsonRpcConnection) {
	var w []string
	for _, gatewayServer := range gateWayServers {
		gatewayServer.loadbalance.methodWebsocketJsonrpcRandConnMap.Range(func(key, value any) bool {
			w = append(w, key.(string))
			return true
		})
	}
	rpcConn.WriteResult(w)
}

func NewGatewayServer() *GatewayServer {
	s := &GatewayServer{
		edgeServerNodeHttpPingTimeoutLimit: 10,
	}
	ws := NewWebsocketServer()
	ws.OnClose(func(rpcConn JsonRpcConnection) {
		s.loadbalance.RemoveByGuid(rpcConn.GUID())
		s.edgeServerNodeWebsocketMap.Delete(rpcConn.GUID())
	})
	s.WebsocketServer = ws
	s.loadAlgorithm = s.loadbalance.LoadRand
	if len(gateWayServers) == 0 {
		InitServerApi()
	}
	gateWayServers = append(gateWayServers, s)
	return s
}

/*
远程RPC调用
同步模式:同步
负载模式:随机
*/
func CallEdgeServerNode(method, token string, requestParams any, responseParams any) (err error) {
	err = errors.New("RPC调用失败,远程网关无法连接")
	request := NewJsonrpcrequest()
	request.Method = method
	request.Params = requestParams
	request.Session = token
	l := len(gateWayServers)
	if l == 0 {
		return
	}
	gateWayServers[rand.Intn(l)].CallEdgeServerNode(request, func(jrc JsonRpcConnection) {
		resp := jrc.GetRpcResponse()
		if resp.Error.Code != 200 {
			err = errors.New(resp.Error.Message)
			return
		}
		err = jrc.ReadResult(responseParams)
	})
	request.Await()
	return
}

func (gs *GatewayServer) SetLoadAlgorithm(loadType int) {
	gs.Lock()
	defer gs.Unlock()
	switch loadType {
	case LOAD_ALGORITHM_RAND:
		gs.loadAlgorithm = gs.loadbalance.LoadRand
	case LOAD_ALGORITHM_PRIORITY:
		gs.loadAlgorithm = gs.loadbalance.LoadPriority
	}
}

func (gs *GatewayServer) SetClusterToken(token string) {
	gs.Lock()
	defer gs.Unlock()
	gs.clusterToken = token
}

func togocdn_http_edge_server_node_reg(rpcConn JsonRpcConnection) {
	var node GatewayEdgeServerNode
	rpcConn.ReadParams(&node)
	var token string = rpcConn.GUID()
	for _, gatewayServer := range gateWayServers {
		if node.EdgeServerNodeConfig.ClusterToken != gatewayServer.clusterToken {
			continue
		}

		switch node.Type {
		case "restful":
			for _, v := range node.EdgeServerNodeConfig.Methods {
				gatewayServer.loadbalance.StoreHttpRestfulMethod(v, rpcConn, node.EdgeServerNodeConfig)
			}
		default:

		}

		//定时删除器
		t := time.NewTimer(time.Second * time.Duration(gatewayServer.edgeServerNodeHttpPingTimeoutLimit))
		go func(t *time.Timer, guid string, g *GatewayServer) {
			g.edgeServerNodeHttpKeepalivedTimer.Store(guid, t)
			<-t.C
			g.loadbalance.RemoveByGuid(guid)
			g.edgeServerNodeHttpKeepalivedTimer.Delete(guid)
		}(t, token, gatewayServer)

	}

	restult := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	rpcConn.WriteResult(restult)

}

func togocdn_websocket_edge_server_node_reg(rpcConn JsonRpcConnection) {
	if !rpcConn.IsConnected() {
		return
	}
	var node GatewayEdgeServerNode
	rpcConn.ReadParams(&node)

	if node.DisableHealthCheck {
		log.Print("客户端要求不进行心跳检测 <- " + rpcConn.GUID())
		rpcConn.DisableHealthCheck()
	}

	node.Guid = rpcConn.GUID()
	node.RemoteAddr = rpcConn.GetRemoteAddr()

	for _, gatewayServer := range gateWayServers {
		if node.EdgeServerNodeConfig.ClusterToken != gatewayServer.clusterToken {
			log.Print("集群token错误 无法注册")
			continue
		}
		gatewayServer.edgeServerNodeWebsocketMap.Store(rpcConn.GUID(), node)
	}

	for _, gatewayServer := range gateWayServers {
		if node.EdgeServerNodeConfig.ClusterToken != gatewayServer.clusterToken {
			continue
		}
		gatewayServer.loadbalance.StoreJsonrpcKeepaliveMethods(node.EdgeServerNodeConfig.Methods, rpcConn, node.EdgeServerNodeConfig.Priority)
	}

	rpcConn.WriteResult("ok")
}

func ping(rpcConn JsonRpcConnection) {
	var params struct {
		Token string `json:"token"`
	}
	rpcConn.ReadParams(&params)

	if params.Token == "" {
		rpcConn.GetRpcResponse().Error.Set(500, "token 不能为空")
		rpcConn.Write()
		return
	}

	var finded bool
	for _, gatewayServer := range gateWayServers {

		t, ok := gatewayServer.edgeServerNodeHttpKeepalivedTimer.Load(params.Token)
		if !ok {
			continue
		} else {
			finded = true
			timer := t.(*time.Timer)
			timer.Reset(time.Second * time.Duration(gatewayServer.edgeServerNodeHttpPingTimeoutLimit))
		}
	}
	if !finded {
		rpcConn.GetRpcResponse().Error.Set(500, "token 不存在或已经失效")
		rpcConn.Write()
		return
	} else {
		rpcConn.WriteResult("ok")
	}
}

func reg(rpcConn JsonRpcConnection) {
	log.Print("边缘节点注册请求 <- " + rpcConn.LinkType() + ":" + rpcConn.GetRemoteAddr())
	switch rpcConn.LinkType() {
	case "http":
		togocdn_http_edge_server_node_reg(rpcConn)
	case "websocket":
		togocdn_websocket_edge_server_node_reg(rpcConn)
	default:
		rpcConn.GetRpcResponse().Error.Set(500, "不支持的协议请求")
		rpcConn.Write()
	}
}

func (gs *GatewayServer) ProxyHttpHandller(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(200)
		return
	}

	es := gs.loadbalance.RandPickRestfulProxy(r)

	if es == nil {
		gs.FrontServer.WebServerHandller(w, r)
	} else {
		proxyHttpEdgeServerNode(w, r, es.HostUrl)
	}
}

// RpcConn代理 (异步模式)
func (gs *GatewayServer) ProxyToEdgeServerNode(sourceRpcConn JsonRpcConnection) {
	go func(sourceRpcConn JsonRpcConnection) {
		edgeServerNodeRpcConn := gs.loadAlgorithm(sourceRpcConn.GetRpcRequest().Method)
		if edgeServerNodeRpcConn == nil {
			//log.Print("不存在的服务:" + sourceRpcConn.GetRpcRequest().Method)
			sourceRpcConn.GetRpcResponse().Error.Set(-32601, "")
			sourceRpcConn.Write()
			return
		}

		//不能删除  防止异步并发情况下 回调 id 不一致
		sourceRequestId := sourceRpcConn.GetRpcRequest().Id

		var wait chan int = make(chan int, 1)
		timeOut := time.NewTimer(time.Second * 600)
		edgeServerNodeRpcConn.Call(sourceRpcConn.GetRpcRequest(), func(jrc JsonRpcConnection) {

			resp := jrc.GetRpcResponse()
			resp.Id = sourceRequestId
			sourceRpcConn.WriteResponse(*resp)
			wait <- 1
		})
		select {
		case <-timeOut.C:
			sourceRpcConn.GetRpcResponse().Error.Set(JSONRPC_408_REQUEST_TIMEOUT, "REQUEST_TIMEOUT")
			sourceRpcConn.Write()
		case <-wait:
		}
	}(sourceRpcConn)

}

// jsonrpc调用边缘节点
func (gs *GatewayServer) CallEdgeServerNode(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	edgeServerNodeRpcConn := gs.loadAlgorithm(rpcRequest.Method)
	if edgeServerNodeRpcConn == nil {
		rpcResponse := NewJsonrpcresponse()
		rpcResponse.Error.Set(502, "502 bad gate way")
		rpc := NewNilRpcConnection(rpcRequest, rpcResponse)
		callback(rpc)
		rpcRequest.Done()
		return
	}
	edgeServerNodeRpcConn.Call(rpcRequest, callback)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}

	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

func proxyHttpEdgeServerNode(w http.ResponseWriter, r *http.Request, targetUrlHost string) {
	target, _ := url.Parse(targetUrlHost)

	targetQuery := target.RawQuery

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	proxy.ServeHTTP(w, r)
}

func getEdgeServerNodeInfo(rpcConn JsonRpcConnection) {

	var result map[string]interface{} = map[string]interface{}{}

	result["restful_api"] = gateWayServers[0].loadbalance.methodHttpRestfulRandConnMap
	result["jsonrpc_api"] = gateWayServers[0].loadbalance.methodWebsocketJsonrpcRandConnMap

	rpcConn.WriteResult(result)
}

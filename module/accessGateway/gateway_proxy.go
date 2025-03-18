package accessGateway

/*
统一接入网关核心算法
by:liangliangit
*/
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/queue"
	"github.com/towgo/towgo/lib/system"

	"github.com/towgo/towgo/module/accountcenter"
	"github.com/towgo/towgo/module/apibilling"
	"github.com/towgo/towgo/module/iolog"
)

var contentTextArr [5]string

func isInContentTextArr(contentType string) bool {
	for _, v := range contentTextArr {
		if v == contentType {
			return true
		}
	}
	return false
}

func init() {
	initContentTextArr()
	iolog.InitManageApi()
	apibilling.InitManageApi()
}

func initContentTextArr() {
	contentTextArr[0] = "text/html"
	contentTextArr[1] = "text/css"
	contentTextArr[2] = "application/javascript"
	contentTextArr[3] = "application/json"
	contentTextArr[4] = "text/plain"
}

type AccessGateWayConfig struct {
	MaxConnectLimit int64
	ListenAdd       string
	ListenPort      string
	ListenPortTLS   string
	TLSCertPath     string
	TLSKeyPath      string
	JsonrpcPath     string
}

type GateWayGroup struct {
	MaxConnectLimit  int64
	Schema           string
	Domain           string
	APIPath          string
	LoadType         string
	UseTargetHost    bool
	IsWebFrontClient bool
	block            chan bool
	funcs            Funcs
	funcsLocker      sync.Mutex
}

func (gwg *GateWayGroup) Run(priority int64) (runToken, doneToken chan bool) {
	runToken = make(chan bool, 1)
	doneToken = make(chan bool, 1)

	fc := &Func{
		Priority:  priority,
		runToken:  runToken,
		doneToken: doneToken,
	}

	gwg.funcsLocker.Lock()
	gwg.funcs = append(gwg.funcs, fc)
	sort.Sort(gwg.funcs)
	gwg.funcsLocker.Unlock()
	gwg.run()
	return
}

func (gwg *GateWayGroup) run() {
	go func(gwg *GateWayGroup) {
		gwg.block <- true   //获取最大连接数令牌
		f := gwg.pickFunc() //获取一条链路
		f.runToken <- true  //分发链路令牌
		<-f.doneToken       //等待链路归还令牌
		<-gwg.block         //归还最大连接数令牌
	}(gwg)
}

// 获取一条链路处理器
func (gwg *GateWayGroup) pickFunc() *Func {
	gwg.funcsLocker.Lock()
	defer gwg.funcsLocker.Unlock()

	if len(gwg.funcs) > 0 {
		f := gwg.funcs[0]
		var newFuncs Funcs
		for k, v := range gwg.funcs {
			if k > 0 {
				newFuncs = append(newFuncs, v)
			}
		}
		gwg.funcs = newFuncs
		return f
	}

	return &Func{}
}

type Funcs []*Func

type Func struct {
	Priority  int64
	doneToken chan bool
	runToken  chan bool
}

// 实现排序len函数
func (m Funcs) Len() int {
	return len(m)
}

func (m Funcs) Less(x, y int) bool {
	return m[x].Priority < m[y].Priority
}

// swap 进行位置置换
func (m Funcs) Swap(x, y int) {
	m[x], m[y] = m[y], m[x]
}

// restful代理
type Method struct {
	TargetHost  string
	ApiRouteWay *ApiRouteWay
	Proxy       *httputil.ReverseProxy
}

type AccessGateWay struct {
	block chan bool
	sync.Mutex
	maxConnectLimit int64
	//proxys         map[string][]*RestfulMethod
	TogoGateWayServer   *jsonrpc.GatewayServer
	restfulGateWayGroup sync.Map
	jsonrpcGateWayGroup sync.Map

	restfulmethods sync.Map //restful
	jsonrpcmethods sync.Map //jsonrpc

	//webclientserver
	WebServerHandller func(http.ResponseWriter, *http.Request)

	//local
	serveMux *http.ServeMux

	//configs
	listenAdd     string
	listenPort    string
	jsonrpcPath   string
	ServerportTLS string
	TLSCertPath   string
	TLSKeyPath    string
}

func NewAccessGateWay(config AccessGateWayConfig) *AccessGateWay {

	if config.MaxConnectLimit <= 0 {
		config.MaxConnectLimit = 0
	}

	gw := &AccessGateWay{
		maxConnectLimit: config.MaxConnectLimit,
		listenAdd:       config.ListenAdd,
		listenPort:      config.ListenPort,
		jsonrpcPath:     config.JsonrpcPath,
		ServerportTLS:   config.ListenPortTLS,
		TLSCertPath:     config.TLSCertPath,
		TLSKeyPath:      config.TLSKeyPath,
		serveMux:        http.NewServeMux(),
	}
	gw.TogoGateWayServer = jsonrpc.NewGatewayServer()
	gw.block = make(chan bool, gw.maxConnectLimit)
	return gw
}

func (hgw *AccessGateWay) LenConnection() int {
	return len(hgw.block)
}

func (hgw *AccessGateWay) Clear() {
	hgw.Lock()
	defer hgw.Unlock()
	hgw.restfulmethods.Range(func(key, _ any) bool {
		hgw.restfulmethods.Delete(key)
		return true
	})

	hgw.jsonrpcmethods.Range(func(key, _ any) bool {
		hgw.jsonrpcmethods.Delete(key)
		return true
	})
}

func (hgw *AccessGateWay) PickRestfulGateWayGroup(domain, serviceUrlPath string) *GateWayGroup {
	proxysAny, ok := hgw.restfulGateWayGroup.Load(domain + serviceUrlPath)
	//完全匹配未命中
	if !ok {
		//模糊匹配
		paths := strings.Split(serviceUrlPath, "/")
		var mathPaths []string
		var path string
		pathsLen := len(paths)

		mathPaths = append(mathPaths, "/")

		for i := 0; i < pathsLen; i++ {
			if paths[i] != "" {
				path = path + "/" + paths[i]
			}

			mathPaths = append(mathPaths, path)
		}

		for i := len(mathPaths) - 1; i >= 0; i-- {
			proxysAny, ok = hgw.restfulGateWayGroup.Load(domain + mathPaths[i])
			if ok {
				break
			}
		}
	}

	//所有匹配都不存在
	if ok {
		return proxysAny.(*GateWayGroup)
	}
	return nil
}

func (hgw *AccessGateWay) PickJsonrpcGateWayGroup(domain, serviceUrlPath string) *GateWayGroup {
	proxysAny, ok := hgw.jsonrpcGateWayGroup.Load(domain + serviceUrlPath)

	//method不存在
	if !ok {
		//模糊匹配
		paths := strings.Split(serviceUrlPath, "/")
		var mathPaths []string
		var path string
		pathsLen := len(paths)

		mathPaths = append(mathPaths, "/")
		for i := 0; i < pathsLen; i++ {
			if paths[i] != "" {
				path = path + "/" + paths[i]
			}
			mathPaths = append(mathPaths, path)
		}

		for i := len(mathPaths) - 1; i >= 0; i-- {
			proxysAny, ok = hgw.jsonrpcGateWayGroup.Load(domain + mathPaths[i])
			if ok {
				break
			}
		}
	}

	//所有匹配都不存在
	if ok {
		if ok {
			return proxysAny.(*GateWayGroup)
		}

	}
	return nil
}

// restful随机路由
func (hgw *AccessGateWay) PickRestfulProxyRand(domain, serviceUrlPath string) *Method {
	proxysAny, ok := hgw.restfulmethods.Load(domain + serviceUrlPath)
	//完全匹配未命中
	if !ok {
		//模糊匹配
		paths := strings.Split(serviceUrlPath, "/")
		var mathPaths []string
		var path string
		pathsLen := len(paths)

		mathPaths = append(mathPaths, "/")

		for i := 0; i < pathsLen; i++ {
			if paths[i] != "" {
				path = path + "/" + paths[i]
			}

			mathPaths = append(mathPaths, path)
		}

		for i := len(mathPaths) - 1; i >= 0; i-- {
			proxysAny, ok = hgw.restfulmethods.Load(domain + mathPaths[i])
			if ok {
				break
			}
		}
	}

	//所有匹配都不存在
	if !ok {
		return nil
	}

	proxys := proxysAny.([]*Method)
	proxysLen := len(proxys)

	if proxysLen == 0 {
		return nil
	}

	if proxysLen == 1 {
		return proxys[0]
	}

	rand.Seed(time.Now().UnixNano())
	restFulMethod := proxys[rand.Intn(proxysLen)]

	return restFulMethod

}

// jsonrpc随机路由
func (hgw *AccessGateWay) PickJsonrpcProxyRand(domain, serviceUrlPath string) *Method {
	proxysAny, ok := hgw.jsonrpcmethods.Load(domain + serviceUrlPath)

	//method不存在
	if !ok {
		//模糊匹配
		paths := strings.Split(serviceUrlPath, "/")
		var mathPaths []string
		var path string
		pathsLen := len(paths)

		mathPaths = append(mathPaths, "/")
		for i := 0; i < pathsLen; i++ {
			if paths[i] != "" {
				path = path + "/" + paths[i]
			}
			mathPaths = append(mathPaths, path)
		}

		for i := len(mathPaths) - 1; i >= 0; i-- {
			proxysAny, ok = hgw.jsonrpcmethods.Load(domain + mathPaths[i])
			if ok {
				break
			}
		}
	}

	//所有匹配都不存在
	if !ok {
		return nil
	}

	proxys := proxysAny.([]*Method)
	proxysLen := len(proxys)

	if proxysLen == 0 {
		return nil
	}

	if proxysLen == 1 {
		return proxys[0]
	}

	jsonrpcMethod := proxys[rand.Intn(proxysLen)]

	return jsonrpcMethod

}

func (hgw *AccessGateWay) GetDomain(hostname string) string {
	host := strings.Split(hostname, ":")
	return host[0]
}

type randRestfulProxyStruct struct {
	Guid   string
	R      *http.Request
	Method *Method
}

// 随机restful代理策略
func randRestfulProxy(hgw *AccessGateWay, w http.ResponseWriter, r *http.Request, isWebFrontClient bool) {
	request_guid := r.Context().Value("guid").(string)

	//账号传输

	method := hgw.PickRestfulProxyRand(hgw.GetDomain(r.Host), r.URL.Path)
	if method != nil {
		//静态代理线路存在 走代理

		//服务接口增加header信息
		r.Header.Del("account")

		requestProxy := func(account accountcenter.Account) {
			apilog := iolog.IoLog{}
			apilog.Guid = request_guid
			apilog.ProxyHost = method.TargetHost
			apilog.Status = iolog.STATUS_WAIT_BACK_RESPONSE
			apilog.Uid = account.ID
			apilog.Username = account.Username
			apilog.ProxyTime = time.Now().Unix()
			queue.NewFuncQueue(func(a ...any) {
				apilog := a[0].(*iolog.IoLog)
				err := basedboperat.Update(apilog, []string{"status", "proxy_time", "proxy_host", "uid", "username"}, "guid = ?", apilog.Guid)
				if err != nil {
					log.Print(err.Error())
				}
			}, &apilog)

			if method.ApiRouteWay.UseApiBilling {
				err := apibilling.Charging(account.ID, r.URL.Path)
				if err != nil {
					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
					return
				}
			}

			rd := randRestfulProxyStruct{
				Guid:   request_guid,
				R:      r,
				Method: method,
			}

			r = r.WithContext(context.WithValue(r.Context(), randRestfulProxyStruct{}, rd))

			//代理异常处理

			method.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {

				apilog := iolog.IoLog{}
				apilog.Guid = request_guid
				if err.Error() == "context canceled" {
					apilog.Status = iolog.STATUS_CLIENT_CANCEL
				} else {
					apilog.Status = iolog.STATUS_PROXY_ERROR
				}
				apilog.Result = err.Error()
				apilog.ResponseTime = time.Now().Unix()
				queue.NewFuncQueue(func(a ...any) {
					apilog := a[0].(*iolog.IoLog)
					basedboperat.Update(apilog, []string{"status", "result", "response_time"}, "guid = ?", apilog.Guid)
				}, &apilog)
				w.WriteHeader(500)
				w.Write([]byte("context canceled"))
			}

			//代理响应重写
			method.Proxy.ModifyResponse = func(w *http.Response) error {
				ctx := w.Request.Context()
				randRestfulProxyStructInterface := ctx.Value(randRestfulProxyStruct{})
				rd := randRestfulProxyStructInterface.(randRestfulProxyStruct)

				b, _ := io.ReadAll(w.Body)

				w.Body = io.NopCloser(bytes.NewReader(b))
				var resultStr string
				if isInContentTextArr(w.Header.Get("Content-Type")) {
					resultStr = string(b)
				}

				queue.NewFuncQueue(func(args ...any) {
					iolog.WriteResult(args[0].(string), args[1].(string))
				},
					rd.Guid, resultStr,
				)

				if rd.Method.ApiRouteWay.UseApiBilling {
					if w.StatusCode < 200 || w.StatusCode >= 300 {
						go apibilling.Refund(account.ID, rd.R.URL.Path)
					}
				}
				return nil
			}
			method.Proxy.ServeHTTP(w, r)
		}

		if isWebFrontClient {
			var account accountcenter.Account
			requestProxy(account)
			return
		}

		_, ok := unauthorizedMethodMap.Load(r.URL.Path + "restful")
		if ok {
			var account accountcenter.Account
			requestProxy(account)
			return
		}

		accountRequest := func(jrc jsonrpc.JsonRpcConnection) {
			responseErr := jrc.GetRpcResponse().Error
			if responseErr.Code != 200 {
				w.WriteHeader(int(responseErr.Code))
				w.Write([]byte(responseErr.Error()))
				queue.NewFuncQueue(func(args ...any) {
					iolog.WriteResult(args[0].(string), args[1].(string))
				},
					request_guid, "未授权",
				)
				return
			}

			var account accountcenter.Account
			jrc.ReadResult(&account)
			if account.ID == 0 {
				w.WriteHeader(401)
				w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
				queue.NewFuncQueue(func(args ...any) {
					iolog.WriteResult(args[0].(string), args[1].(string))
				},
					request_guid, "未授权",
				)
				return
			}
			b, _ := json.Marshal(account)
			r.Header.Set("account", string(b))
			requestProxy(account)
		}

		session := r.Header.Get("session")
		rpcRequest := jsonrpc.NewJsonrpcrequest()
		rpcRequest.Session = session
		rpcRequest.Method = "/account/myinfo"
		hgw.TogoGateWayServer.CallEdgeServerNode(rpcRequest, accountRequest)
		rpcRequest.Await()

	} else {
		//本地客户端服务
		if hgw.WebServerHandller != nil {
			hgw.WebServerHandller(w, r)
		} else {
			//代理线路不存在 返回502
			apilog := iolog.IoLog{}
			apilog.Guid = request_guid
			apilog.Status = iolog.STATUS_PROXY_ERROR
			apilog.Result = "后端代理线路不存在 "
			queue.NewFuncQueue(func(a ...any) {
				apilog := a[0].(*iolog.IoLog)
				basedboperat.Update(apilog, []string{"status", "proxy_host"}, "guid = ?", apilog.Guid)
			}, &apilog)

			w.WriteHeader(502)
			w.Write([]byte("access gateway 502 Bad Gateway"))
		}
	}

}

func proxyTo(hgw *AccessGateWay, rpcConn jsonrpc.JsonRpcConnection, r *http.Request) {
	rpcRequest := rpcConn.GetRpcRequest()
	method := hgw.PickJsonrpcProxyRand(hgw.GetDomain(r.Host), rpcRequest.Method)
	if method != nil {
		//静态代理线路存在 走代理
		rpcClient := jsonrpc.NewHttpClient()
		callback := func(rpcResult jsonrpc.Jsonrpcresponse) {
			rpcResponse := rpcConn.GetRpcResponse()
			rpcResponse.Error = rpcResult.Error
			rpcResponse.Timestampin = rpcResult.Timestampin
			rpcResponse.Timestampout = rpcResult.Timestampout
			rpcResponse.Result = rpcResult.Result
			rpcConn.Write()
		}
		rpcClient.ErrorFunc = func(err error) {
			rpcConn.GetRpcResponse().Error.Set(-1, "502 bad gate way")
			rpcConn.GetRpcResponse().Error.NewChild(500, err.Error())
			rpcConn.Write()
		}
		rpcClient.Call(method.TargetHost+hgw.jsonrpcPath, rpcRequest, callback)

	} else {
		//websocket 自动化负载均衡网关
		//检查本地服务
		if jsonrpc.HasMethod(rpcConn.GetRpcRequest().Method) {
			jsonrpc.Exec(rpcConn)
		} else {
			hgw.TogoGateWayServer.ProxyToEdgeServerNode(rpcConn)
			rpcRequest.Await()
		}

	}
}

// 随机jsonrpc代理策略
func randJsonrpcProxy(hgw *AccessGateWay, w http.ResponseWriter, r *http.Request) {

	rpcConn := jsonrpc.NewHttpRpcConnection(w, r)
	if rpcConn == nil {
		return
	}

	rpcConn.GetRpcRequest().Route.SourceAddr = system.RemoteIp(r)

	rpcRequest := rpcConn.GetRpcRequest()
	if rpcRequest.Method == "" {
		rpcConn.GetRpcResponse().Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	//账号传输
	rpcRequest.Ctx = map[jsonrpc.ContextKey]interface{}{}

	accountRequest := func(jrc jsonrpc.JsonRpcConnection) {

		responseErr := jrc.GetRpcResponse().Error
		if responseErr.Code != 200 {
			rpcConn.GetRpcResponse().Error.Set(responseErr.Code, responseErr.Message)
			rpcConn.Write()
			return
		}

		var account accountcenter.Account
		jrc.ReadResult(&account)
		if account.ID == 0 {
			rpcConn.GetRpcResponse().Error.Set(401, "未授权")
			rpcConn.Write()
			return
		}

		//非超级管理员权限 进行接口权限验证
		if !account.HasJurisdiction("superadmin") {
			//验证接口权限
			if !account.HasJurisdiction(rpcRequest.Method) {
				rpcConn.GetRpcResponse().Error.Set(403, "无访问权限")
				rpcConn.Write()
				return
			}
		}

		rpcConn.GetRpcRequest().Ctx["account"] = account

		proxyTo(hgw, rpcConn, r)
	}

	_, ok := unauthorizedMethodMap.Load(rpcRequest.Method + "jsonrpc")
	if ok {
		//未验证允许放行
		proxyTo(hgw, rpcConn, r)
	} else {
		rpcRequestCall := jsonrpc.NewJsonrpcrequest()
		rpcRequestCall.Session = rpcRequest.Session
		rpcRequestCall.Method = "/account/myinfo"
		hgw.TogoGateWayServer.CallEdgeServerNode(rpcRequestCall, accountRequest)

		<-rpcConn.GetRpcRequest().Context().Done()
	}

}

// restful 统一入口
func (hgw *AccessGateWay) RequestHandlerForRestfuls(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(200)
		return
	}

	remoteIp := system.RemoteIp(r)
	r.Header.Set("X-Forwarded-For", remoteIp)
	r.Header.Set("X-Real-IP", remoteIp)
	r.Header.Set("Ali-CDN-Real-IP", remoteIp)

	guid := system.GetGUID().Hex()
	apilog := iolog.IoLog{}
	apilog.Guid = guid
	apilog.Status = iolog.STATUS_QUEUE
	apilog.RemoteHost = r.RemoteAddr
	apilog.RequestApiPath = r.URL.Path
	apilog.RequestTime = time.Now().Unix()
	apilog.RequestUrl = hgw.GetDomain(r.Host)

	queue.NewFuncQueue(func(a ...any) {
		basedboperat.Create(a[0].(*iolog.IoLog))
	}, &apilog)

	//有并发限制 进入并发队列
	if hgw.maxConnectLimit > 0 {
		hgw.block <- true
		defer func() {
			<-hgw.block
		}()
	}

	//路由组

	group := hgw.PickRestfulGateWayGroup(hgw.GetDomain(r.Host), r.URL.Path)
	if group != nil { //组策略命中， 优先级最高
		priorityStr := r.URL.Query().Get("accessGateWayPriority")
		var priority int64 = 0
		if priorityStr != "" {
			p, err := strconv.ParseInt(priorityStr, 10, 8)
			if err != nil {
				priority = 0
			} else {
				priority = p
			}
		}
		runToken, doneToken := group.Run(priority)
		<-runToken //等待运行令牌
		time.Sleep(time.Millisecond * 5)
		b, _ := io.ReadAll(r.Body)
		time.Sleep(time.Millisecond * 5)

		select {
		case <-r.Context().Done():
			log.Print("客户端请求超时")
			apilog := iolog.IoLog{}
			apilog.Guid = guid
			apilog.Status = iolog.STATUS_CLIENT_CANCEL
			queue.NewFuncQueue(func(a ...any) {
				apilog := a[0].(*iolog.IoLog)
				basedboperat.Update(apilog, []string{"status"}, "guid = ?", apilog.Guid)
			}, &apilog)
			doneToken <- true
			return
		default:
			ctx := context.WithValue(r.Context(), "guid", guid)

			r = r.WithContext(ctx)
			r.Body = io.NopCloser(bytes.NewReader(b))
			randRestfulProxy(hgw, w, r, group.IsWebFrontClient) //运行任务 //随机方案
			doneToken <- true
			//归还运行令牌
			return
		}
	} else {
		ctx := context.WithValue(r.Context(), "guid", guid)
		//var context context.WithValue(contcontext.back)
		r = r.WithContext(ctx)
		randRestfulProxy(hgw, w, r, false) //随机方案
	}

}

// jsonrpc统一入口
func (hgw *AccessGateWay) RequestHandlerForJsonrpc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}
	//总路由并发连接数
	if hgw.maxConnectLimit > 0 {
		hgw.block <- true
		defer func(hgw *AccessGateWay) {
			<-hgw.block
		}(hgw)
	}

	//路由组
	group := hgw.PickJsonrpcGateWayGroup(hgw.GetDomain(r.Host), r.URL.Path)
	if group != nil { //组策略命中， 优先级最高
		if group.IsWebFrontClient {
			hgw.RequestHandlerForRestfuls(w, r)
			return
		}

		priorityStr := r.URL.Query().Get("accessGateWayPriority")
		var priority int64 = 0
		if priorityStr != "" {
			p, err := strconv.ParseInt(priorityStr, 10, 8)
			if err != nil {
				priority = 0
			} else {
				priority = p
			}
		}

		w.Header().Set("Content-Type", "application/json")

		//单链路并发连接数
		runToken, doneToken := group.Run(priority)
		<-runToken                  //等待运行令牌
		randJsonrpcProxy(hgw, w, r) //运行任务	//随机方案
		doneToken <- true           //归还运行令牌
		return
	} else {

		w.Header().Set("Content-Type", "application/json")
		randJsonrpcProxy(hgw, w, r) //随机方案
	}

}

func (hgw *AccessGateWay) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	hgw.serveMux.HandleFunc(pattern, handler)
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

// 新增代理组 maxConnectlimit must > 0
func (hgw *AccessGateWay) NewGroupProxy(maxConnectlimit int64, domain, serviceUrlPath, APIType, loadType string, isWebFrontClient bool) error {
	hgw.Lock()
	defer hgw.Unlock()
	key := domain + serviceUrlPath

	if maxConnectlimit < 0 {
		maxConnectlimit = 0
	}

	group := &GateWayGroup{
		MaxConnectLimit:  maxConnectlimit,
		LoadType:         loadType,
		APIPath:          serviceUrlPath,
		IsWebFrontClient: isWebFrontClient,
		block:            make(chan bool, maxConnectlimit),
	}
	switch APIType {
	case "restful":
		_, ok := hgw.restfulGateWayGroup.Load(key)
		if ok {
			return errors.New("目标组已经存在")
		}
		hgw.restfulGateWayGroup.Store(key, group)
	case "jsonrpc":
		_, ok := hgw.jsonrpcGateWayGroup.Load(key)
		if ok {
			return errors.New("目标组已经存在")
		}
		hgw.jsonrpcGateWayGroup.Store(key, group)
	default:
		return errors.New("APIType error only eg: restful | jsonrpc")
	}
	return nil
}

// 删除代理组
func (hgw *AccessGateWay) RemoveGroupProxy(domain, serviceUrlPath, APIType string) error {
	hgw.Lock()
	defer hgw.Unlock()
	key := domain + serviceUrlPath
	switch APIType {
	case "restful":
		hgw.restfulGateWayGroup.Delete(key)
	case "jsonrpc":
		hgw.jsonrpcGateWayGroup.Delete(key)
	default:
		return errors.New("APIType error only eg: restful | jsonrpc")
	}
	return nil
}

// 新增restful代理
func (hgw *AccessGateWay) NewRestfulProxy(group ApiRouteGroup, apiRouteWay *ApiRouteWay) error {
	hgw.Lock()
	defer hgw.Unlock()
	targetHost := apiRouteWay.GetTargetHostUrl()
	target, err := url.Parse(targetHost)

	targetQuery := target.RawQuery

	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		//使用目标主机头去请求（过滤源信息）
		if group.UseTargetHostToRequest {
			req.Host = target.Host
		}

		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if apiRouteWay.RemoveProxyPath {
			req.URL.Path = strings.Replace(req.URL.Path, group.APIPath, "", 1)
		}
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

	restfulMethod := &Method{
		TargetHost:  targetHost,
		ApiRouteWay: apiRouteWay,
		Proxy:       proxy,
	}
	//proxys := []*RestfulMethod{restfulMethod}
	proxysAny, loaded := hgw.restfulmethods.Load(group.Domain + group.APIPath)
	if loaded {
		proxys := proxysAny.([]*Method)
		if isTargetHostInProxys(targetHost, proxys) {
			return errors.New(targetHost + ":目标地址已经存在（重复）")
		}
		//加载到集群列表中
		proxys = append(proxys, restfulMethod)
		hgw.restfulmethods.Store(group.Domain+group.APIPath, proxys)
	} else {
		proxys := []*Method{restfulMethod}
		hgw.restfulmethods.Store(group.Domain+group.APIPath, proxys)
	}
	return nil
}

// 移除restful代理
func (hgw *AccessGateWay) RemoveRestfulProxy(domain string, serviceUrlPath, targetHost string) error {
	hgw.Lock()
	defer hgw.Unlock()
	proxysAny, loaded := hgw.restfulmethods.LoadAndDelete(domain + serviceUrlPath)
	if loaded {
		proxys := proxysAny.([]*Method)
		proxys = removeTargetHostInProxys(targetHost, proxys)
		if len(proxys) == 0 {
			hgw.restfulmethods.Delete(domain + serviceUrlPath)
		} else {
			hgw.restfulmethods.Store(domain+serviceUrlPath, proxys)
		}
	}
	return nil
}

// 新增jsonrpc代理
func (hgw *AccessGateWay) NewJsonrpcProxy(domain, serviceUrlPath string, apiRouteWay *ApiRouteWay) error {
	hgw.Lock()
	defer hgw.Unlock()
	targetHost := apiRouteWay.GetTargetHostUrl()
	url, err := url.Parse(targetHost)

	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	restfulMethod := &Method{
		TargetHost:  targetHost,
		ApiRouteWay: apiRouteWay,
		Proxy:       proxy,
	}
	//proxys := []*RestfulMethod{restfulMethod}
	proxysAny, loaded := hgw.jsonrpcmethods.Load(domain + serviceUrlPath)
	if loaded {
		proxys := proxysAny.([]*Method)
		if isTargetHostInProxys(targetHost, proxys) {
			return errors.New("jsonrpc method 目标地址已经存在（重复）")
		}
		//加载到集群列表中
		proxys = append(proxys, restfulMethod)
		hgw.jsonrpcmethods.Store(domain+serviceUrlPath, proxys)
	} else {
		proxys := []*Method{restfulMethod}
		hgw.jsonrpcmethods.Store(domain+serviceUrlPath, proxys)
	}
	return nil
}

// 移除jsonrpc代理
func (hgw *AccessGateWay) RemoveJsonrpcProxy(domain, serviceUrlPath, targetHost string) error {
	hgw.Lock()
	defer hgw.Unlock()
	proxysAny, loaded := hgw.jsonrpcmethods.Load(domain + serviceUrlPath)
	if loaded {
		proxys := proxysAny.([]*Method)
		proxys = removeTargetHostInProxys(targetHost, proxys)
		if len(proxys) == 0 {
			hgw.jsonrpcmethods.Delete(domain + serviceUrlPath)
		} else {
			hgw.jsonrpcmethods.Store(domain+serviceUrlPath, proxys)
		}
	}
	return nil
}

// 目标主机是否在列表中
func isTargetHostInProxys(targetHost string, proxys []*Method) bool {
	for _, v := range proxys {
		if targetHost == v.TargetHost {
			return true
		}
	}
	return false
}

func removeTargetHostInProxys(targetHost string, proxys []*Method) []*Method {
	var tmpProxys []*Method
	for _, v := range proxys {
		if targetHost != v.TargetHost {
			tmpProxys = append(tmpProxys, v)
		}
	}
	return tmpProxys
}

func (hgw *AccessGateWay) ListenAndServe() {
	hgw.serveMux.HandleFunc("/", hgw.RequestHandlerForRestfuls)
	hgw.serveMux.HandleFunc(hgw.jsonrpcPath, hgw.RequestHandlerForJsonrpc)
	hgw.TogoGateWayServer.SetLoadAlgorithm(jsonrpc.LOAD_ALGORITHM_RAND)
	hgw.serveMux.HandleFunc("/websocket"+hgw.jsonrpcPath, hgw.TogoGateWayServer.WebsocketServer.WebsocketServiceHandller.ServeHTTP)

	if hgw.ServerportTLS != "" {
		go func() {
			err := http.ListenAndServeTLS(hgw.listenAdd+":"+hgw.ServerportTLS, hgw.TLSCertPath, hgw.TLSKeyPath, hgw.serveMux)
			if err != nil {
				log.Print(err.Error())
			}
		}()
	}

	http.ListenAndServe(hgw.listenAdd+":"+hgw.listenPort, hgw.serveMux)
}

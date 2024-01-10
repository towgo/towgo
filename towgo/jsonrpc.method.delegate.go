/*
json rpc 2.0 方法委托
by:liangliangit
version 2.1
*/
package towgo

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/towgo/towgo/lib/system"
)

// 委托任务列表
// var funcs map[string]func(JsonRpcConnection) = map[string]func(JsonRpcConnection){}
var lock sync.Mutex
var funcs map[string]func(JsonRpcConnection) = map[string]func(JsonRpcConnection){}
var lockedMethods sync.Map

var BeforExec func(rpcConn JsonRpcConnection)
var AfterExec func(rpcConn JsonRpcConnection)
var OnMethodNotFound func(rpcConn JsonRpcConnection)

// 查询method是否存在
func HasMethod(method string) bool {
	_, ok := funcs[method]
	return ok
}

// 为所有method增加头
func AddMethodHead(methodHead string) {
	lock.Lock()
	defer lock.Unlock()
	var newmap map[string]func(JsonRpcConnection) = map[string]func(JsonRpcConnection){}
	for k, v := range funcs {
		newmap[methodHead+k] = v
	}
	funcs = newmap
}

// 获取method列表
func GetMethods() (method []string) {
	for k, _ := range funcs {
		method = append(method, k)
	}
	return
}

func http_jsonrpc_wrapper(w http.ResponseWriter, r *http.Request) {

	urlPath := r.URL.Path
	rpcRequest := NewJsonrpcrequest()
	rpcRequest.Method = urlPath
	rpcRequest.Session = r.Header.Get("session")

	conn := &HttpRpcConnection{
		guid:        "HTTP:" + system.GetGUID().Hex(),
		response:    w,
		request:     r,
		rpcRequest:  rpcRequest,
		rpcResponse: NewJsonrpcresponse(),
		httpwrapper: true,
	}

	err := defaultJsonRpcInterceptor(conn)
	if err != nil {
		conn.isConnected = false
		log.Print(err.Error())
		return //拦截后 rpc响应由拦截器处理，  不需要再次响应
	}
	Exec(conn)
}

func MethodToHttpPathInterface(serveMux *http.ServeMux) {
	for k, _ := range funcs {
		method := "/" + strings.TrimLeft(k, "/")
		serveMux.HandleFunc(method, http_jsonrpc_wrapper)
	}
}

// 锁定指定method （可用于许可证到期锁定相关服务）
func MethodLock(method string) {
	lockedMethods.Store(method, "")
}

// 解锁method
func MethodUnlock(method string) {
	lockedMethods.Delete(method)
}

// 锁定所有method （可用于许可证到期锁定相关服务,排除关键性业务不锁定）
func MethodLockAll(excludeMethods ...string) {
	for k, _ := range funcs {
		lockedMethods.Store(k, "")
	}
	for _, v := range excludeMethods {
		lockedMethods.Delete(v)
	}
}

// 解锁所有method
func MethodUnlockAll(excludeMethods ...string) {
	lockedMethods.Range(func(key, _ any) bool {
		lockedMethods.Delete(key)
		return true
	})
	for _, v := range excludeMethods {
		lockedMethods.Store(v, "")
	}
}

// 设定委托任务
func SetFunc(method string, f func(JsonRpcConnection)) {
	lock.Lock()
	defer lock.Unlock()
	funcs[method] = f
}

// 委托执行任务
func Exec(rpcConn JsonRpcConnection) {
	if BeforExec != nil {
		BeforExec(rpcConn)
	}
	rpcResponse := rpcConn.GetRpcResponse()
	rpcRequest := rpcConn.GetRpcRequest()
	if rpcRequest == nil {
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	if rpcRequest.Method == "" {
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	//判断是否锁定
	_, ok := lockedMethods.Load(rpcRequest.Method)
	if ok {
		rpcResponse.Error.Set(500, "Method has been locked!")
		rpcConn.Write()
		return
	}

	methodFunc, exists := funcs[rpcRequest.Method]
	// 判断委托是否存在
	if !exists {
		//如果注册了Method not found处理函数  不进行默认响应
		if OnMethodNotFound != nil {
			OnMethodNotFound(rpcConn)
			return
		}
		//默认响应 Method not found找不到方法
		rpcResponse.Error.Set(-32601, "")
		rpcConn.Write()
		return
	}

	// 执行委托的程序
	methodFunc(rpcConn)
	if AfterExec != nil {
		AfterExec(rpcConn)
	}
}

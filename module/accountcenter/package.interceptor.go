package accountcenter

import (
	"errors"
	"sync"

	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/requestblocker"
)

var unauthorizedMethod map[string]bool = map[string]bool{}

var methodLocker sync.Mutex
var noInterceptor bool

var loginBlocker *requestblocker.RequestBlocker

func init() {

	//每秒最大请求登录、注册数  /超过会被拦截
	loginBlocker = requestblocker.New(1, 100)
}

//账户登录次数过多请求拦截器 end

func initInterceptor() {
	//未认证的授权
	unauthorizedMethod["/account/login"] = true
	unauthorizedMethod["/account/logoff"] = true
	unauthorizedMethod["/account/reg"] = true
	unauthorizedMethod["ping"] = true
	unauthorizedMethod["pong"] = true

	jsonrpc.AddInterceptor(accountInterceptor)

}

func NoInterceptor(b bool) {
	noInterceptor = b
}

func UnauthorizedMethodAdd(method string) {
	methodLocker.Lock()
	unauthorizedMethod[method] = true
	methodLocker.Unlock()
}

func UnauthorizedMethodDel(method string) {
	methodLocker.Lock()
	delete(unauthorizedMethod, method)
	methodLocker.Unlock()
}

// 账户中心拦截器
func accountInterceptor(conn jsonrpc.JsonRpcConnection) error {

	if noInterceptor {
		return nil
	}

	var err error

	rpcRequest := conn.GetRpcRequest()
	rpcResponse := conn.GetRpcResponse()

	var account *Account
	account, err = account.LoginByToken(rpcRequest.Session)
	if err != nil {
		if !unauthorizedMethod[rpcRequest.Method] {
			rpcResponse.Error.Set(401, err.Error())
			conn.Write()
			return errors.New("requestMethod:" + rpcRequest.Method + ":未登录,无法访问")
		} else {
			return nil
		}
	}

	account.Password = ""

	rpcRequest.Ctx = map[jsonrpc.ContextKey]interface{}{"account": account}

	return nil

}

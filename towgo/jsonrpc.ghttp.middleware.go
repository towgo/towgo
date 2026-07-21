package towgo

import "github.com/gogf/gf/v2/net/ghttp"

type ghttpJsonRpcContextKey string

const ghttpJsonRpcConnectionKey ghttpJsonRpcContextKey = "towgo/jsonrpc/connection"

func JsonRpcRequest(r *ghttp.Request) {
	w := r.Response.Writer
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")

	rpcConn := NewHttpRpcConnection(w, r.Request)
	if rpcConn == nil {
		return
	}
	rpcConn.isConnected = true
	rpcConn.WithContext(r.GetCtx())
	r.SetCtxVar(ghttpJsonRpcConnectionKey, rpcConn)
	r.Middleware.Next()
}

func JsonRpcResponse(r *ghttp.Request) {
	r.Middleware.Next()

	rpcConn := GhttpJsonRpcConnection(r)
	if rpcConn == nil {
		return
	}
	rpcConn.finish()
}

func GhttpJsonRpcConnection(r *ghttp.Request) *HttpRpcConnection {
	v := r.GetCtxVar(ghttpJsonRpcConnectionKey).Val()
	if rpcConn, ok := v.(*HttpRpcConnection); ok {
		return rpcConn
	}
	return nil
}

func GhttpMiddlewareHandler(r *ghttp.Request) {
}

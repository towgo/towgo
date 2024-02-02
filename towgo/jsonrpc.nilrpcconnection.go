package towgo

import (
	"sync"

	"github.com/towgo/towgo/lib/system"
)

type NilRpcConnection struct {
	guid        string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	sync.Map
}

func (w *NilRpcConnection) SetValue(key string, value any) {
	w.Store(key, value)
}

func (w *NilRpcConnection) GetValue(key string) (value any, ok bool) {
	return w.Load(key)
}

func (n *NilRpcConnection) WriteError(code int64, msg string) {
	n.rpcResponse.Error.Set(code, msg)
	n.Write()
}

// 获取远程客户端IP
func (n *NilRpcConnection) GetRemoteAddr() string {
	return ""
}

// 读取原生的json字符串
func (n *NilRpcConnection) Read() string {
	return ""
}

// 将参数映射到传入的指针
func (n *NilRpcConnection) ReadParams(destParams ...interface{}) error {
	return nil
}

// 将结果映射到传入的指针
func (n *NilRpcConnection) ReadResult(destResult ...interface{}) error {
	return nil
}

// 将结果写入（最终会组装成一个响应对象发送给对端）
func (n *NilRpcConnection) WriteResult(any interface{}) {

}

// 写入连接器内置的响应对象
func (n *NilRpcConnection) Write() {

}

// 直接将传入的响应对象写入
func (n *NilRpcConnection) WriteResponse(Jsonrpcresponse) {

}

// 获取连接器请求对象
func (n *NilRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return n.rpcRequest
}

// 获取连接器响应对象
func (n *NilRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return n.rpcResponse
}

/*
推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
*/
func (n *NilRpcConnection) Push(request *Jsonrpcrequest) error {
	return nil
}

/*
全双工模式下可以作为客户端进行请求通讯
*/
func (n *NilRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {

}

// 连接器的底层协议类型 tcp|http|websocket
func (n *NilRpcConnection) LinkType() string {
	return "nil"
}

func (n *NilRpcConnection) IsConnected() bool {
	return false
}

func (n *NilRpcConnection) GUID() string {
	return n.guid
}

func (n *NilRpcConnection) Close() {

}

func NewNilRpcConnection(rpcRequest *Jsonrpcrequest, rpcResponse *Jsonrpcresponse) JsonRpcConnection {
	return &NilRpcConnection{
		rpcRequest:  rpcRequest,
		rpcResponse: rpcResponse,
		guid:        system.GetGUID().Hex(),
	}
}
func (c *NilRpcConnection) EnableHealthCheck() {
}

func (c *NilRpcConnection) DisableHealthCheck() {
}

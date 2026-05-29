package towgo

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/towgo/towgo/v2/lib/system"
)

type LocalRpcConnection struct {
	guid        string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	ctx         context.Context
	nextFunc    func()
	err         error
	result      interface{}
	sync.Map
}

func (w *LocalRpcConnection) SetValue(key string, value any) {
	w.Store(key, value)
}

func (w *LocalRpcConnection) GetValue(key string) (value any, ok bool) {
	return w.Load(key)
}

func (n *LocalRpcConnection) WriteError(code int64, msg string) {
	n.rpcResponse.Error.Set(code, msg)
	n.Write()
}

// 获取远程客户端IP
func (n *LocalRpcConnection) GetRemoteAddr() string {
	return "local"
}

// 读取原生的json字符串
func (n *LocalRpcConnection) Read() string {
	b, _ := json.Marshal(n.rpcRequest)
	return string(b)
}

// 将参数映射到传入的指针
func (n *LocalRpcConnection) ReadParams(destParams ...interface{}) error {
	b, _ := json.Marshal(n.rpcRequest.Params)
	for _, v := range destParams {
		err := json.Unmarshal(b, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// 将结果映射到传入的指针
func (n *LocalRpcConnection) ReadResult(destResult ...interface{}) error {
	b, _ := json.Marshal(n.rpcResponse.Result)
	for _, v := range destResult {
		err := json.Unmarshal(b, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *LocalRpcConnection) SetResult(result interface{}) {
	n.result = result
}

func (n *LocalRpcConnection) GetResult() interface{} {
	return n.result
}

// 将结果写入（最终会组装成一个响应对象发送给对端）
func (n *LocalRpcConnection) WriteResult(any interface{}) {
	n.result = any
	n.Write()
}

// 写入响应
func (n *LocalRpcConnection) Write() {
	if n.rpcResponse == nil {
		n.rpcResponse = NewJsonrpcresponse()
	}
	defer n.rpcRequest.Done()
	n.rpcResponse.Id = n.rpcRequest.Id
	n.rpcResponse.Timestampin = n.rpcRequest.Timestampin
	time := time.Now().UnixNano() / 1e6
	n.rpcResponse.Timestampout = strconv.FormatInt(time, 10)

	// 设置 result 和 error
	if n.err != nil {
		n.rpcResponse.Error.Set(500, n.err.Error())
	} else if n.result != nil {
		n.rpcResponse.Result = n.result
	}
}

// 直接将传入的响应对象写入
func (n *LocalRpcConnection) WriteResponse(Jsonrpcresponse) {
	// 空实现
}

// 获取连接器请求对象
func (n *LocalRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return n.rpcRequest
}

// 获取连接器响应对象
func (n *LocalRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return n.rpcResponse
}

/*
推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
*/
func (n *LocalRpcConnection) Push(request *Jsonrpcrequest) error {
	return nil
}

/*
全双工模式下可以作为客户端进行请求通讯
*/
func (n *LocalRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {

	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	conn := NewLocalRpcConnection(rpcRequest, nil)

	Exec(conn)
	rpcRequest.Done()
	callback(conn)
}

// 连接器的底层协议类型 tcp|http|websocket
func (n *LocalRpcConnection) LinkType() string {
	return "local"
}

func (n *LocalRpcConnection) IsConnected() bool {
	return true
}

func (n *LocalRpcConnection) GUID() string {
	return n.guid
}

func (n *LocalRpcConnection) Close() {

}

func NewLocalRpcConnection(rpcRequest *Jsonrpcrequest, rpcResponse *Jsonrpcresponse) JsonRpcConnection {
	if rpcResponse == nil {
		rpcResponse = NewJsonrpcresponse()
	}
	return &LocalRpcConnection{
		rpcRequest:  rpcRequest,
		rpcResponse: rpcResponse,
		guid:        system.GetGUID().Hex(),
	}
}

func (c *LocalRpcConnection) EnableHealthCheck() {
}

func (c *LocalRpcConnection) DisableHealthCheck() {
}

func (c *LocalRpcConnection) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

func (c *LocalRpcConnection) WithContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *LocalRpcConnection) GetError() error {
	return c.err
}

func (c *LocalRpcConnection) WithError(err error) {
	c.err = err
}

func (c *LocalRpcConnection) Next() {
	if c.nextFunc != nil {
		c.nextFunc()
	}
}

func (c *LocalRpcConnection) SetNextFunc(fn func()) {
	c.nextFunc = fn
}

/*
JSON-RPC2.0 over HTTP for golang
by:liangliangit
ver 1.0
*/
package towgo

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/towgo/towgo/lib/system"
)

type HttpRpcConnection struct {
	guid        string
	isConnected bool
	remoteAddr  string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	response    http.ResponseWriter
	request     *http.Request
	bodyStr     string
	paramsBytes []byte
	resultBytes []byte
}

func (c *HttpRpcConnection) WriteError(code int64, msg string) {
	c.rpcResponse.Error.Set(code, msg)
	c.Write()
}

func (c *HttpRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.rpcRequest
}
func (c *HttpRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.rpcResponse
}

func (c *HttpRpcConnection) Write() {
	defer c.rpcRequest.ctxCancel()
	if c.rpcResponse.Id == "" {
		c.rpcResponse.Id = c.rpcRequest.Id
	}
	c.rpcResponse.Timestampin = c.rpcRequest.Timestampin
	time := time.Now().UnixNano() / 1e6
	c.rpcResponse.Timestampout = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(c.rpcResponse)
	if c.rpcRequest.Isencryption {
		if isencryption {
			code, _ := AesEncrypt(mjson)
			mjson = code
		}
	}
	c.response.Write(mjson)
}

func (c *HttpRpcConnection) WriteResult(result interface{}) {
	c.rpcResponse.Result = result
	c.Write()
}

func (c *HttpRpcConnection) WriteResponse(rpcResponse Jsonrpcresponse) {
	c.rpcResponse = &rpcResponse
	c.Write()
}

func (c *HttpRpcConnection) Read() string {
	if c.bodyStr != "" {
		return c.bodyStr
	}
	data, _ := io.ReadAll(c.request.Body)
	datastr := string(data)
	c.bodyStr = datastr
	return c.bodyStr
}

// 读取参数
func (c *HttpRpcConnection) ReadParams(destParams ...interface{}) error {
	if len(c.paramsBytes) == 0 {
		var err error
		c.paramsBytes, err = json.Marshal(c.rpcRequest.Params)
		if err != nil {
			return err
		}
	}
	for _, v := range destParams {
		err := json.Unmarshal(c.paramsBytes, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// 读取结果
func (c *HttpRpcConnection) ReadResult(destParams ...interface{}) error {

	if len(c.resultBytes) == 0 {
		var err error
		c.resultBytes, err = json.Marshal(c.rpcResponse.Result)
		if err != nil {
			return err
		}
	}
	for _, v := range destParams {
		err := json.Unmarshal(c.resultBytes, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// 获取对方ip地址
func (c *HttpRpcConnection) GetRemoteAddr() string {
	if c.remoteAddr != "" {
		return c.remoteAddr
	}
	c.remoteAddr = system.RemoteIp(c.request)
	return c.remoteAddr
}

/*
推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
*/
func (c *HttpRpcConnection) Push(request *Jsonrpcrequest) error {
	if request.Method == "" {
		return errors.New("method not set")
	}
	if request.Id == "" {
		request.Id = system.RandChar(64)
	}
	time := time.Now().UnixNano() / 1e6
	request.Timestampin = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(request)
	if c.rpcRequest.Isencryption {
		if isencryption {
			code, _ := AesEncrypt(mjson)
			mjson = []byte(code)
		}
	}
	_, err := c.response.Write(mjson)
	return err
}

// 连接类型
func (c *HttpRpcConnection) LinkType() string {
	return "http"
}

func NewHttpRpcConnection(w http.ResponseWriter, r *http.Request) *HttpRpcConnection {

	conn := &HttpRpcConnection{
		guid:        "HTTP:" + system.GetGUID().Hex(),
		response:    w,
		request:     r,
		rpcResponse: NewJsonrpcresponse(),
	}
	bodyStr := conn.Read()
	if bodyStr == "" {
		return nil
	}

	rpcRequest, err := ToJsonrpcrequest(bodyStr)

	if rpcRequest.Isencryption {
		b, _ := json.Marshal(rpcRequest)
		conn.bodyStr = string(b)
	}

	conn.rpcResponse.Id = rpcRequest.Id

	if err != nil {
		log.Print(err.Error())
	}

	conn.rpcRequest = rpcRequest

	return conn
}

// http协议的服务端不支持call方法（因为http是短链接 无法进行全双工通讯）
func (wsc *HttpRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	log.Print("http protocol is not support call function")
}

// http 全局唯一入口 包装器 将http请求包装成jsonrpc请求
func HttpHandller(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")

	rpcConn := NewHttpRpcConnection(w, r)

	if rpcConn == nil {
		return
	}
	rpcConn.isConnected = true

	//运行拦截器
	err := defaultJsonRpcInterceptor(rpcConn)
	if err != nil {
		rpcConn.isConnected = false
		log.Print(err.Error())
		return //拦截后 rpc响应由拦截器处理，  不需要再次响应
	}
	//未被拦截 调用rpc方法
	Exec(rpcConn)
	rpcConn.isConnected = false //http是断链接  调用完RPC后默认连接关闭
}

func (c *HttpRpcConnection) IsConnected() bool {
	return c.isConnected
}

func (c *HttpRpcConnection) GUID() string {
	return c.guid
}

func (c *HttpRpcConnection) EnableHealthCheck() {
	log.Print("http protocol is not support EnableHealthCheck function")
}

func (c *HttpRpcConnection) DisableHealthCheck() {
	log.Print("http protocol is not support DisableHealthCheck function")
}

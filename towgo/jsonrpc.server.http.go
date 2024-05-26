/*
JSON-RPC2.0 over HTTP for golang
by:liangliangit
ver 1.0
*/
package towgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
)

type HttpRpcConnection struct {
	guid        string
	httpwrapper bool
	isConnected bool
	remoteAddr  string
	rpcRequest  *Jsonrpcrequest
	rpcResponse *Jsonrpcresponse
	response    http.ResponseWriter
	request     *http.Request
	bodyStr     string
	paramsBytes []byte
	resultBytes []byte
	sync.Map
}

func (c *HttpRpcConnection) SetValue(key string, value any) {
	c.Store(key, value)
}

func (c *HttpRpcConnection) GetValue(key string) (value any, ok bool) {
	return c.Load(key)
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

	var mjson []byte

	if c.httpwrapper {
		if c.rpcResponse.Error.Code == 200 {
			mjson, _ = json.Marshal(c.rpcResponse.Result)
		} else {
			mjson, _ = json.Marshal(c.rpcResponse.Error)
		}

	} else {
		mjson, _ = json.Marshal(c.rpcResponse)
	}

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

	if c.httpwrapper {
		r := c.request
		switch r.Method {
		case "POST":
			if len(c.paramsBytes) == 0 {
				var err error
				c.paramsBytes, err = io.ReadAll(r.Body)
				if err != nil {
					return err
				}
			}
		case "GET":
			for _, v := range destParams {

				myStruct, err := parseQueryParams(r, reflect.TypeOf(v))
				if err != nil {
					return err
				}

				// 将结构体转换为JSON
				jsonData, err := json.Marshal(myStruct)
				if err != nil {
					return err
				}

				err = json.Unmarshal(jsonData, v)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if len(c.paramsBytes) == 0 {
		var err error
		c.paramsBytes, err = json.Marshal(c.rpcRequest.Params)
		if err != nil {
			log.Print(err.Error())
		}
	}

	for _, v := range destParams {
		err := json.Unmarshal(c.paramsBytes, v)
		if err != nil {
			log.Print(err.Error())
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
			log.Print(err.Error())
		}
	}
	for _, v := range destParams {
		err := json.Unmarshal(c.resultBytes, v)
		if err != nil {
			log.Print(err.Error())
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

	/*
		defer func(rpcConn JsonRpcConnection) {
			r := recover()
			if r != nil {
				// 处理其他panic异常
				var errors []error
				// 捕获第一条panic异常
				if err, ok := r.(error); ok {
					errors = append(errors, err)
				}
				for {
					if r = recover(); r == nil {
						break
					}

					if err, ok := r.(error); ok {
						errors = append(errors, err)
					}
				}

				// 打印错误信息
				fmt.Println("发生以下错误：")
				for _, err := range errors {
					fmt.Println(err)
				}
				rpcConn.WriteError(500, DEFAULT_ERROR_MSG)
			}
		}(rpcConn)
	*/

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

func (c *HttpRpcConnection) Close() {

}

func (c *HttpRpcConnection) EnableHealthCheck() {
	log.Print("http protocol is not support EnableHealthCheck function")
}

func (c *HttpRpcConnection) DisableHealthCheck() {
	log.Print("http protocol is not support DisableHealthCheck function")
}

func parseQueryParams(r *http.Request, t reflect.Type) (interface{}, error) {
	queryValues := r.URL.Query()
	for {
		if t.Kind().String() == "ptr" {
			t = t.Elem()
		} else {
			break
		}
	}

	structValue := reflect.New(t).Elem()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name
		fieldValue := structValue.FieldByName(fieldName)

		queryParam := queryValues.Get(fieldName)
		if queryParam == "" {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(queryParam)
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(queryParam)
			if err != nil {
				return nil, fmt.Errorf("failed to parse bool value for field '%s'", fieldName)
			}
			fieldValue.SetBool(boolValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intValue, err := strconv.ParseInt(queryParam, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse int value for field '%s'", fieldName)
			}
			fieldValue.SetInt(intValue)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintValue, err := strconv.ParseUint(queryParam, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse uint value for field '%s'", fieldName)
			}
			fieldValue.SetUint(uintValue)
		// 处理其他类型...
		default:
			return nil, fmt.Errorf("unsupported field type for field '%s'", fieldName)
		}
	}

	return structValue.Interface(), nil
}

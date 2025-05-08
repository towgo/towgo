/*
json rpc 2.0 模型
by:liangliangit
*/
package towgo

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

type ContextKey string

const (
	SESSION                         = "jsonrpc.session"
	JSON_RPC_CONNECTION_CONTEXT_KEY = "json_rpc_connection"
	DEFAULT_ERROR_MSG               = "网络开小差了"
)

var isencryption bool

func SetSecretkey(key string, iv string) {
	if key == "" {
		isencryption = false
		return
	}
	isencryption = true
	SetIv(iv)
	SetKey(key)
}

// jsonrpc请求
type Jsonrpcrequest struct {
	sync.Mutex
	ctx          context.Context            `json:"-"`
	ctxCancel    context.CancelFunc         `json:"-"`
	Jsonrpc      string                     `json:"jsonrpc"`
	Method       string                     `json:"method"`
	DataType     string                     `json:"DataType"`
	Params       interface{}                `json:"params"`
	Id           string                     `json:"id"`
	Ctx          map[ContextKey]interface{} `json:"ctx"`
	Session      string                     `json:"session"`
	Timestampin  string                     `json:"timestampin"`
	Timestampout string                     `json:"timestampout"`
	Isencryption bool                       `json:"-"`
	Route        Route                      `json:"route"`
}

// jsonrpc响应
type Jsonrpcresponse struct {
	Jsonrpc      string                     `json:"jsonrpc"`
	DataType     string                     `json:"DataType"`
	Result       interface{}                `json:"result"`
	Id           string                     `json:"id"`
	Ctx          map[ContextKey]interface{} `json:"ctx"`
	Timestampin  string                     `json:"timestampin"`
	Timestampout string                     `json:"timestampout"`
	Route        Route                      `json:"route"`
	Error        Error                      `json:"error"`
}

type Jsonrpcresponseclient struct {
	Jsonrpc      string                 `json:"jsonrpc"`
	Result       interface{}            `json:"result"`
	Id           string                 `json:"id"`
	Ctx          map[string]interface{} `json:"ctx"`
	Timestampin  string                 `json:"timestampin"`
	Timestampout string                 `json:"timestampout"`
	Route        Route                  `json:"route"`
	Error        interface{}            `json:"error"`
}

type Route struct {
	TTL        int64
	DestAddr   string
	SourceAddr string
}

// 新建一个响应
func NewJsonrpcresponse() *Jsonrpcresponse {
	return &Jsonrpcresponse{
		Jsonrpc: "2.0",
		Error: Error{
			Code:    200,
			Message: "ok",
		},
	}
}

func NewJsonrpcrequest() *Jsonrpcrequest {
	ctx, cancel := context.WithCancel(context.Background())
	return &Jsonrpcrequest{
		Jsonrpc:   "2.0",
		ctx:       ctx,
		ctxCancel: cancel,
	}
}

// 请求包含上下文
func (j *Jsonrpcrequest) WithContext(ctx context.Context) {
	j.Lock()
	defer j.Unlock()
	j.ctx = ctx
}

// 请求完成
func (j *Jsonrpcrequest) Done() {
	j.ctxCancel()
}

/*
异步情况下  使用await 来等待响应 。
例如一个http的jsonrpc请求 后端又是websocket的异步请求，此时如果http层不等待
那么，http层就会出现未等待到websocket的数据返回前就关闭了客户端的连接连接，导致客户端数据丢失
*/

func (j *Jsonrpcrequest) Await() {
	<-j.Context().Done()
}

func (j *Jsonrpcrequest) Context() context.Context {
	return j.ctx
}

// 将json字符串转换成一个请求对象
func ToJsonrpcrequest(s string) (*Jsonrpcrequest, error) {

	jsonrpcrequest := NewJsonrpcrequest()

	//尝试正常转换
	e := json.Unmarshal([]byte(s), jsonrpcrequest)

	if e != nil {
		//log.Print(e.Error())
		//如果加密开启
		if isencryption {

			b, e := AesDecrypt(s)

			if e != nil {
				log.Print(e.Error())
				return jsonrpcrequest, e
			}

			e = json.Unmarshal(b, jsonrpcrequest)
			if e != nil {
				log.Print(e.Error())
			}

			jsonrpcrequest.Isencryption = true
			if jsonrpcrequest.DataType != "" {
				jsonrpcrequest.Params = ProcessSourceData(jsonrpcrequest.DataType, s, true)
			}
			return jsonrpcrequest, nil
			//解密
		} else {
			return jsonrpcrequest, e
		}

	} else {
		if jsonrpcrequest.DataType != "" {
			jsonrpcrequest.Params = ProcessSourceData(jsonrpcrequest.DataType, s, true)
		}
		return jsonrpcrequest, e
	}
}

// 将json字符串转换成一个响应对象
func ToJsonrpcresponse(s string) (Jsonrpcresponse, error) {

	var jsonrpcresponse Jsonrpcresponse = Jsonrpcresponse{}

	e := json.Unmarshal([]byte(s), &jsonrpcresponse)

	if e != nil {

		if isencryption {

			b, e := AesDecrypt(s)

			if e != nil {
				log.Print(e.Error())
				return jsonrpcresponse, e
			}
			e = json.Unmarshal([]byte(b), &jsonrpcresponse)
			if e != nil {
				log.Print(e.Error())
			}
			if jsonrpcresponse.DataType != "" {
				jsonrpcresponse.Result = ProcessSourceData(jsonrpcresponse.DataType, s, false)
			}
			return jsonrpcresponse, e
			//解密
		} else {
			return jsonrpcresponse, e
		}

	} else {
		if jsonrpcresponse.DataType != "" {
			jsonrpcresponse.Result = ProcessSourceData(jsonrpcresponse.DataType, s, false)
		}
		return jsonrpcresponse, e
	}
}

func ProcessSourceData(dataType string, jsonStr string, isRrequest bool) interface{} {

	var data struct {
		Params []byte `json:"params"`
		Result []byte `json:"result"`
	}
	switch dataType {
	case "[]byte":
		json.Unmarshal([]byte(jsonStr), &data)

	}
	if isRrequest {
		return data.Params
	} else {
		return data.Result
	}
}

func ToJsonrpcresponseclient(s string) (Jsonrpcresponseclient, error) {

	var jsonrpcresponse Jsonrpcresponseclient

	e := json.Unmarshal([]byte(s), &jsonrpcresponse)

	if e == nil {
		return jsonrpcresponse, e
	} else {
		return Jsonrpcresponseclient{}, e
	}
}

func (jrr *Jsonrpcresponse) ReadResult(destParams interface{}) error {

	paramsBytes, err := json.Marshal(jrr.Result)
	if err != nil {
		return err
	}

	return json.Unmarshal(paramsBytes, destParams)
}

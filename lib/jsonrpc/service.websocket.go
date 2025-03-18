/*
JSON-RPC2.0 over websocket
by:liangliangit
websocket连接
长连接、可降低http短连接带来的额外tcp握手开销，减少TIME_WAIT过多导致的端口被全部占用问题
推荐内部服务使用
*/
package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
	"golang.org/x/net/websocket"
)

var DefaultWebSocketServer *WebSocketServer
var clientCallTimeOut int64 = 600

type WebSocketServer struct {
	instantRequest           int64
	regOnConnectFuncLock     sync.Mutex
	regOnCloseFuncLock       sync.Mutex
	wsOnConnectCallFuncs     []func(rpcConn JsonRpcConnection)
	wsOnCloseCallFuncs       []func(rpcConn JsonRpcConnection)
	WebsocketServiceHandller websocket.Handler
	healthCheckCancel        *sync.Map
}

type WebSocketRpcConnection struct {
	lock                   sync.Mutex
	guid                   string
	isConnected            bool
	messageChan            chan string
	CallTimeOut            int64 //客户端请求后，等待响应的超时时间
	PingDuringTime         int64 //ping时间间隔
	requestBody            string
	request                *Jsonrpcrequest
	response               *Jsonrpcresponse
	wsConn                 *websocket.Conn
	paramsBytes            []byte
	rpcCallBackFuncs       *sync.Map
	requestCallBackCancels *sync.Map
	healCheckCancel        context.CancelFunc
}

type RpcCallback struct {
	Request  *Jsonrpcrequest
	Callback func(JsonRpcConnection)
}

func init() {
	DefaultWebSocketServer = NewWebsocketServer()
}

func SetClientCallTimeOut(second int64) {
	clientCallTimeOut = second
}

func NewWebsocketServer() *WebSocketServer {
	w := &WebSocketServer{
		healthCheckCancel: &sync.Map{},
	}
	w.WebsocketServiceHandller = w.preHandller
	return w
}

func (w *WebSocketRpcConnection) WriteError(code int64, msg string) {
	w.response.Error.Set(code, msg)
	w.Write()
}

func (w *WebSocketRpcConnection) Duplicate() (new *WebSocketRpcConnection) {
	new = NewWebSocketRpcConnection(w.wsConn)
	new.guid = w.guid
	new.lock = w.lock
	new.messageChan = w.messageChan
	new.isConnected = w.isConnected
	new.rpcCallBackFuncs = w.rpcCallBackFuncs
	new.requestCallBackCancels = w.requestCallBackCancels
	new.healCheckCancel = w.healCheckCancel
	return
}

func (w *WebSocketServer) preHandller(ws *websocket.Conn) {

	ws.MaxPayloadBytes = 1024 * 1024 * 100 //最大传输数据
	rpcConn := NewWebSocketRpcConnection(ws)
	rpcConn.isConnected = true
	w.onConnect(rpcConn)

	//websocket获取数据
	go func(rpcConn *WebSocketRpcConnection, w *WebSocketServer) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("err=%v , stack=%s\n", err, debug.Stack())
			}
		}()
		var err error
		for {
			var message string
			if err = websocket.Message.Receive(ws, &message); err != nil {
				log.Println("websocket receive error:", err.Error())
				rpcConn.Close()
				break
			}
			w.instantRequest++
			rpcConn.messageChan <- message
		}
	}(rpcConn, w)

	for {
		message, ok := rpcConn.ReceiveMessage()
		if !ok {
			log.Print(rpcConn.GetRemoteAddr() + ":通道关闭")
			break
		}

		//复制主控rpcConnect 防止业务层使用并发而导致主控被更改
		tmpRpcConn := rpcConn.Duplicate()
		tmpRpcConn.AnalysisByString(message)

		request := tmpRpcConn.GetRpcRequest()
		if request.Method == "" {
			rpcResponse := tmpRpcConn.GetRpcResponse()
			if rpcResponse.Id == "" {
				log.Print("websocket 收到无效 jsonrpc response 信息 -> " + message)
				continue
			}

			//rpc响应
			rpcCallbackInterface, ok := rpcConn.rpcCallBackFuncs.Load(rpcResponse.Id)
			if ok {
				go func(rpcCallbackInterface any, tmpRpcConn *WebSocketRpcConnection) {
					defer func() {
						if err := recover(); err != nil {
							log.Printf("err=%v , stack=%s\n", err, debug.Stack())
						}
					}()
					rpcCallback := rpcCallbackInterface.(*RpcCallback)
					tmpRpcConn.request = rpcCallback.Request
					rpcCallback.Callback(tmpRpcConn)
					tmpRpcConn.request.Done()
				}(rpcCallbackInterface, tmpRpcConn)

			}
			cancel, ok := rpcConn.requestCallBackCancels.Load(rpcResponse.Id)
			if ok {
				c := cancel.(context.CancelFunc)
				c()
			}
		} else { //rpc请求

			//委托任务
			go func(tmpRpcConn *WebSocketRpcConnection) {
				defer func(tmpRpcConn *WebSocketRpcConnection) {
					if err := recover(); err != nil {
						log.Printf("err=%v , stack=%s\n", err, debug.Stack())
						tmpRpcConn.WriteError(500, DEFAULT_ERROR_MSG)
						tmpRpcConn.request.Done()
					}

				}(tmpRpcConn)
				err := defaultJsonRpcInterceptor(tmpRpcConn)
				if err != nil {
					log.Print(fmt.Errorf("请求被拦截:%w", err))
					return //拦截后 rpc响应由拦截器处理，  不需要再次响应
				}
				Exec(tmpRpcConn)

				//结束ctx
				tmpRpcConn.request.Done()
			}(tmpRpcConn)

		}
	}
	rpcConn.isConnected = false
	w.onClose(rpcConn)
}

func (w *WebSocketServer) OnConnect(callback func(rpcConn JsonRpcConnection)) {
	w.regOnConnectFuncLock.Lock()
	defer w.regOnConnectFuncLock.Unlock()
	w.wsOnConnectCallFuncs = append(w.wsOnConnectCallFuncs, callback)
}

func (w *WebSocketServer) OnClose(callback func(rpcConn JsonRpcConnection)) {
	w.regOnCloseFuncLock.Lock()
	defer w.regOnCloseFuncLock.Unlock()
	w.wsOnCloseCallFuncs = append(w.wsOnCloseCallFuncs, callback)
}

func (w *WebSocketServer) onConnect(rpcConn *WebSocketRpcConnection) {
	log.Print("ws client On OnConnect : " + rpcConn.GUID())
	rpcConn.EnableHealthCheck()
	//通知注册函数
	for _, v := range w.wsOnConnectCallFuncs {
		v(rpcConn)
	}
}

// 心跳检测
func (wsrc *WebSocketRpcConnection) EnableHealthCheck() {
	if wsrc.healCheckCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	wsrc.healCheckCancel = cancel
	go func(ctx context.Context, rpcConn *WebSocketRpcConnection) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				request := NewJsonrpcrequest()
				request.Method = "ping"
				var hasResponse bool
				rpcConn.Call(request, func(jrc JsonRpcConnection) {
					hasResponse = true
				})

				//指定的时间后没有收到响应 ， 认为对方已经断线， 关闭socket连接
				if wsrc.PingDuringTime == 0 {
					wsrc.PingDuringTime = 5
				}
				time.Sleep(time.Second * time.Duration(wsrc.PingDuringTime))
				if !hasResponse {
					log.Print(rpcConn.GUID() + " -> 链路失去响应,主动断开连接")

					rpcConn.Close()

					return
				}
			}
		}
	}(ctx, wsrc)
}

// 关闭心跳检测
func (wsrc *WebSocketRpcConnection) DisableHealthCheck() {
	if wsrc.healCheckCancel != nil {
		wsrc.healCheckCancel()
		wsrc.healCheckCancel = nil
	}
}

func (w *WebSocketServer) onClose(rpcConn *WebSocketRpcConnection) {

	log.Print("ws client On Close : " + rpcConn.GUID())
	rpcConn.DisableHealthCheck()
	for _, v := range w.wsOnCloseCallFuncs {
		v(rpcConn)
	}
}

func NewWebSocketRpcConnection(ws *websocket.Conn) *WebSocketRpcConnection {
	return &WebSocketRpcConnection{
		guid:                   "WS:" + system.GetGUID().Hex(),
		messageChan:            make(chan string, 1),
		CallTimeOut:            clientCallTimeOut,
		wsConn:                 ws,
		response:               NewJsonrpcresponse(),
		rpcCallBackFuncs:       &sync.Map{},
		requestCallBackCancels: &sync.Map{},
	}
}

// 解析消息
func (w *WebSocketRpcConnection) AnalysisByString(message string) {
	//rpc 请求
	w.requestBody = message
	rpcRequest, _ := ToJsonrpcrequest(message)
	w.request = rpcRequest
	w.paramsBytes = nil

	if rpcRequest.Method == "" {
		response, _ := ToJsonrpcresponse(message)
		w.response = &response
	} else {
		w.response = NewJsonrpcresponse()
	}
}

// 获取远程客户端IP
func (w *WebSocketRpcConnection) GetRemoteAddr() string {
	if w.request != nil {
		if w.request.Route.SourceAddr != "" {
			return w.request.Route.SourceAddr
		}
	}
	if w.wsConn != nil {
		return w.wsConn.RemoteAddr().String()
	}
	return ""
}

// 读取
func (w *WebSocketRpcConnection) Read() string {
	return w.requestBody
}

func (w *WebSocketRpcConnection) ReadParams(destParams ...interface{}) error {
	if len(w.paramsBytes) == 0 {
		var err error
		w.paramsBytes, err = json.Marshal(w.request.Params)
		if err != nil {
			return err
		}
	}
	for _, v := range destParams {
		err := json.Unmarshal(w.paramsBytes, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WebSocketRpcConnection) ReadResult(destResult ...interface{}) error {
	result := w.response.Result
	b, _ := json.Marshal(result)
	for _, v := range destResult {
		err := json.Unmarshal(b, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *WebSocketRpcConnection) WriteResult(any interface{}) {
	w.response.Result = any
	w.Write()
}

func (w *WebSocketRpcConnection) Write() {
	defer w.request.Done()
	w.response.Id = w.request.Id
	w.response.Timestampin = w.request.Timestampin
	time := time.Now().UnixNano() / 1e6
	w.response.Timestampout = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(w.response)
	if w.request.Isencryption {
		if isencryption {
			code, _ := AesEncrypt(mjson)
			mjson = code
		}
	}
	websocket.Message.Send(w.wsConn, string(mjson))
}

func (w *WebSocketRpcConnection) writeResponse() {
	defer w.request.Done()
	w.response.Timestampin = w.request.Timestampin
	time := time.Now().UnixNano() / 1e6
	w.response.Timestampout = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(w.response)
	if w.request.Isencryption {
		if isencryption {
			code, _ := AesEncrypt(mjson)
			mjson = code
		}
	}
	websocket.Message.Send(w.wsConn, string(mjson))
}

func (w *WebSocketRpcConnection) WriteResponse(response Jsonrpcresponse) {
	w.response = &response
	w.writeResponse()
}

func (w *WebSocketRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return w.request
}

func (w *WebSocketRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return w.response
}

/*
推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
*/
func (w *WebSocketRpcConnection) Push(request *Jsonrpcrequest) error {
	if request.Method == "" {
		return errors.New("method not set")
	}

	if request.Id == "" {
		request.Id = system.RandChar(64)
	}
	time := time.Now().UnixNano() / 1e6
	request.Timestampin = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(request)
	if w.request.Isencryption {
		if isencryption {
			code, _ := AesEncrypt(mjson)
			mjson = []byte(code)
		}
	}
	return websocket.Message.Send(w.wsConn, string(mjson))
}

// 连接类型
func (*WebSocketRpcConnection) LinkType() string {
	return "websocket"
}

func (wsc *WebSocketRpcConnection) MockResponse(response *Jsonrpcresponse) {
	if !wsc.IsConnected() {
		return
	}
	b, _ := json.Marshal(response)
	wsc.messageChan <- string(b)
}

func (wsc *WebSocketRpcConnection) ReceiveMessage() (string, bool) {
	if !wsc.IsConnected() {
		return "", false
	}
	msg, ok := <-wsc.messageChan
	return msg, ok
}

func (wsc *WebSocketRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {

	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}

	rpcCallBack := &RpcCallback{
		Request:  rpcRequest,
		Callback: callback,
	}

	wsc.rpcCallBackFuncs.Store(rpcRequest.Id, rpcCallBack)
	ctx, cancel := context.WithCancel(context.Background())
	wsc.requestCallBackCancels.Store(rpcRequest.Id, cancel)

	go wsc.requestDoneOrTimeOut(ctx, rpcRequest.Id)
	if !wsc.IsConnected() {
		resp := NewJsonrpcresponse()
		resp.Error.Set(500, "websocket没有连接")
		wsc.response = resp
		callback(wsc)
		return
	}
	_, err = wsc.wsConn.Write(bytesData)
	if err != nil {
		resp := NewJsonrpcresponse()
		resp.Error.Set(500, err.Error())
		wsc.response = resp
		callback(wsc)
		return
	}

	//delete requestCallBackFunc by requestid

}

func (wsc *WebSocketRpcConnection) requestDoneOrTimeOut(ctx context.Context, requestid string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("err=%v , stack=%s\n", err, debug.Stack())
		}
	}()
	t := time.NewTimer(time.Second * time.Duration(wsc.CallTimeOut))
	select {
	case <-ctx.Done():
		t.Stop()
	case <-t.C:
		if wsc == nil {
			return
		}

		resp := NewJsonrpcresponse()
		resp.Id = requestid
		resp.Error.Set(JSONRPC_408_REQUEST_TIMEOUT, "REQUEST_TIMEOUT")
		wsc.MockResponse(resp)
	}
	wsc.rpcCallBackFuncs.Delete(requestid)
	wsc.requestCallBackCancels.Delete(requestid)
}

func (wsc *WebSocketRpcConnection) IsConnected() bool {
	return wsc.isConnected
}

func (wsc *WebSocketRpcConnection) GUID() string {
	return wsc.guid
}

// 关闭连接
func (wsc *WebSocketRpcConnection) Close() {
	if wsc == nil {
		log.Print("wsc == nill")
		return
	}

	wsc.lock.Lock()
	defer wsc.lock.Unlock()
	if wsc.isConnected {
		wsc.isConnected = false
		close(wsc.messageChan)
		wsc.wsConn.Close()
	}
}

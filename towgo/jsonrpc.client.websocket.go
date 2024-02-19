/*
json rpc 2.0 websocket客户端
by:liangliangit
websocket连接
长连接、可降低http短连接带来的额外tcp握手开销，减少TIME_WAIT过多导致的端口被全部占用问题
推荐内部服务使用
*/
package towgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
	"golang.org/x/net/websocket"
)

type WebScoketClient struct {
	CallTimeOut            int64 //客户端请求后，等待响应的超时时间
	url                    string
	origin                 string
	autoReConnect          bool
	conn                   *websocket.Conn
	rpcCallBackFuncs       sync.Map
	requestCallBackCancels sync.Map
	OnConnect              func(*WebScoketClient)
	OnClose                func(*WebScoketClient)
	healCheckCancel        context.CancelFunc
	pingTimeOut            int64
}

// 关闭心跳检测
func (wsc *WebScoketClient) DisableHealthCheck() {
	if wsc.healCheckCancel != nil {
		wsc.healCheckCancel()
		wsc.healCheckCancel = nil
	}
}

func (wsc *WebScoketClient) AutoReConnect(connect bool) {
	wsc.autoReConnect = connect
}

// 心跳检测
func (wsc *WebScoketClient) EnableHealthCheck() {
	if wsc.healCheckCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	wsc.healCheckCancel = cancel
	go func(ctx context.Context, rpcConn *WebScoketClient) {
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
				time.Sleep(time.Second * time.Duration(wsc.pingTimeOut))
				if !hasResponse {
					log.Print("服务端链路失去响应,主动断开连接", " timeout ", wsc.pingTimeOut, " second")
					rpcConn.Close()
					return
				}
			}
		}
	}(ctx, wsc)
}

func NewWebsocketClient(url, origin string) *WebScoketClient {

	wsc := &WebScoketClient{
		CallTimeOut:   600,
		url:           url,
		origin:        origin,
		autoReConnect: true,
		pingTimeOut:   10,
	}

	return wsc
}

func (wsc *WebScoketClient) SetPingTimeOut(second int64) {
	wsc.pingTimeOut = second
}

func (wsc *WebScoketClient) Close() {
	wsc.conn.Close()
}

func (wsc *WebScoketClient) onClose() {
	wsc.DisableHealthCheck()
	if wsc.OnClose != nil {
		go wsc.OnClose(wsc)
	}
	if wsc.autoReConnect {
		go func() {
			time.Sleep(time.Second * 1)
			wsc.reConnect()
		}()
	}
}

func (wsc *WebScoketClient) Connect() {
	wsc.reConnect()
}

func (wsc *WebScoketClient) ReLoad(url, origin string) {
	wsc.url = url
	wsc.origin = origin
	autoReConnect := wsc.autoReConnect
	if autoReConnect {
		wsc.autoReConnect = false
	}
	if wsc.conn != nil {
		wsc.conn.Close()
	}

	wsc.autoReConnect = autoReConnect
	conn, err := websocket.Dial(wsc.url, "", wsc.origin)
	if err != nil {
		log.Print(err.Error())
		wsc.onClose()
		return
	}
	wsc.conn = conn

	go clientConnHandler(wsc)
}

func (wsc *WebScoketClient) reConnect() {
	conn, err := websocket.Dial(wsc.url, "", wsc.origin)
	if err != nil {
		log.Print(err.Error())
		wsc.onClose()
		return
	}
	conn.MaxPayloadBytes = 1024 * 1024 * 100 //最大传输1G数据
	wsc.conn = conn

	go clientConnHandler(wsc)
}

func clientConnHandler(wsc *WebScoketClient) {
	if wsc.OnConnect != nil {
		go wsc.OnConnect(wsc)

	}
	wsc.EnableHealthCheck()
	rpcConn := NewWebSocketRpcConnection(wsc.conn)
	var err error
	for {
		var message string
		//BIO
		if err = websocket.Message.Receive(wsc.conn, &message); err != nil {
			break
		}
		tmpRpcConn := rpcConn.Duplicate()
		tmpRpcConn.AnalysisByString(message)

		request := tmpRpcConn.GetRpcRequest()
		if request.Method == "" {
			rpcResponse := tmpRpcConn.GetRpcResponse()
			if rpcResponse.Id == "" {
				//无效信息
				log.Print("websocket 收到无效 jsonrpc response 信息")
				continue
			}
			//rpc响应
			rpcCallbackInterface, ok := wsc.rpcCallBackFuncs.Load(rpcResponse.Id)
			if ok {

				go func(rpcCallbackInterface any, tmpRpcConn *WebSocketRpcConnection) {
					defer func() {
						if err := recover(); err != nil {
							log.Printf("error: %s\n", err)
						}
					}()
					rpcCallback := rpcCallbackInterface.(*RpcCallback)
					tmpRpcConn.request = rpcCallback.Request
					rpcCallback.Callback(tmpRpcConn)
					tmpRpcConn.request.Done()
				}(rpcCallbackInterface, tmpRpcConn)

			}
			cancel, ok := wsc.requestCallBackCancels.Load(rpcResponse.Id)
			if ok {
				c := cancel.(context.CancelFunc)
				c()
			}

		} else { //rpc请求
			//委托任务
			go func(tmpRpcConn *WebSocketRpcConnection) {
				defer func(tmpRpcConn *WebSocketRpcConnection) {
					err := recover()
					if err != nil {
						log.Print(err)
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
	wsc.onClose()
}

func (wsc *WebScoketClient) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	if rpcRequest.Id == "" {
		rpcRequest.Id = system.GetGUID().Hex()
	}

	if wsc.conn == nil {
		return
	}

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

	wsc.conn.Write(bytesData)

	//delete requestCallBackFunc by requestid
	go wsc.requestDoneOrTimeOut(ctx, rpcRequest.Id)
}

// 请求超时或完成后的清理逻辑
func (wsc *WebScoketClient) requestDoneOrTimeOut(ctx context.Context, requestid string) {
	t := time.NewTimer(time.Second * time.Duration(wsc.CallTimeOut))
	select {
	case <-ctx.Done():
		t.Stop()
		wsc.rpcCallBackFuncs.Delete(requestid)
	case <-t.C:
		rpcCallbackInterface, ok := wsc.rpcCallBackFuncs.LoadAndDelete(requestid)
		if ok {
			rpcCallback := rpcCallbackInterface.(*RpcCallback)
			request := rpcCallback.Request
			response := NewJsonrpcresponse()
			response.Error.Set(JSONRPC_408_REQUEST_TIMEOUT, "jsonrpc request time out")
			rpcConn := NewNilRpcConnection(request, response)
			rpcCallback.Callback(rpcConn)
			request.Done()
		}
	}
	wsc.requestCallBackCancels.Delete(requestid)
}

// push 是不需要响应求的 可以作为事件通知
func (wsc *WebScoketClient) Push(rpcRequest *Jsonrpcrequest) {
	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}
	wsc.conn.Write(bytesData)
}

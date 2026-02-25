/*
json rpc 2.0 tcp客户端
by:liangliangit
*/
package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
)

type TcpClient struct {
	connected              bool
	OnMessage              func(rpcConn JsonRpcConnection)
	OnClose                func(tcpClient *TcpClient)
	OnConnect              func(tcpClient *TcpClient)
	requestCallBackFuncs   sync.Map
	requestCallBackCancels sync.Map
	timeoutInterval        int64
	ip                     string
	port                   string
	ctx                    context.Context
	PingInterval           int64
	*TcpConn
	cancel      context.CancelFunc
	CallTimeOut int64
}

// 客户端数据拆分
func (tcpClient *TcpClient) tcpClientHandller(tcpConn *TcpConn, data string) {

	rpcConn := NewTcpRpcConnection(tcpConn.conn, data)

	tcpClient.OnMessage(rpcConn)
	//如果method信息为空那么判断这条消息是响应
	if rpcConn.GetRpcRequest().Method == "" {

		rpcResponse := rpcConn.GetRpcResponse()

		if rpcResponse.Id != "" {
			//rpc响应
			f, ok := tcpClient.requestCallBackFuncs.Load(rpcResponse.Id)
			if ok {
				callback := f.(func(JsonRpcConnection))
				callback(rpcConn)
			}
			cancel, ok := tcpClient.requestCallBackCancels.Load(rpcResponse.Id)
			if ok {
				c := cancel.(context.CancelFunc)
				c()
			}

		}
	} else {
		Exec(rpcConn)
	}
}

func NewTcpClient(ip, port string) (*TcpClient, error) {

	tcpConn := &TcpConn{
		PingInterval: 20,
		BUFFERLENGTH: BUFFERLENGTH, //内存缓冲区长度 （字节） 1024000byte=100MB
		DATAEND:      DATAEND,      //数据尾帧标识符
	}

	tcpClient := &TcpClient{
		ip:              ip,
		port:            port,
		TcpConn:         tcpConn,
		PingInterval:    20,
		timeoutInterval: 60,
		CallTimeOut:     clientCallTimeOut,
	}
	return tcpClient, nil
}

func (tcpClient *TcpClient) IsConnected() bool {
	return tcpClient.connected
}

func (tcpClient *TcpClient) Connect() error {
	if tcpClient.IsConnected() {
		return errors.New("已经连接")
	}
	conn, err := net.Dial("tcp", tcpClient.ip+":"+tcpClient.port)
	if err != nil {
		tcpClient.connected = false
		tcpClient.OnClose(tcpClient)
		return err
	}

	tcpClient.connected = true

	tcpclientctx, cancel := context.WithCancel(context.Background())
	tcpClient.cancel = cancel
	tcpClient.ctx = tcpclientctx
	tcpClient.TcpConn.conn = conn

	//开启独立读取数据线路（全双工模式）
	go ConnReadingClient(tcpClient.ctx, tcpClient, tcpClient.TcpConn, tcpClient.tcpClientHandller)

	tcpClient.OnConnect(tcpClient)
	return nil
}

// 关闭连接
func (tcpClient *TcpClient) Close() {
	//关闭tcp连接
	tcpClient.TcpConn.conn.Close()
	//关闭通知上下文通道关闭
	if tcpClient.TcpConn != nil {
		if tcpClient.TcpConn.cancel != nil {
			tcpClient.TcpConn.cancel()
		}
	}

	if tcpClient.cancel != nil {
		tcpClient.cancel()
	}

}

func (tcpClient *TcpClient) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}

	tcpClient.requestCallBackFuncs.Store(rpcRequest.Id, callback)

	ctx, cancel := context.WithCancel(context.Background())
	tcpClient.requestCallBackCancels.Store(rpcRequest.Id, cancel)

	//delete requestCallBackFunc by requestid
	go func(ctx context.Context, requestid string) {
		t := time.NewTimer(time.Second * time.Duration(tcpClient.CallTimeOut))
		select {
		case <-ctx.Done():
			t.Stop()
		case <-t.C:
		}
		tcpClient.requestCallBackFuncs.Delete(requestid)
		tcpClient.requestCallBackCancels.Delete(requestid)
	}(ctx, rpcRequest.Id)

	//发送数据
	bytes := []byte{tcpClient.DATAEND}
	bytesData = append(bytesData, bytes...)
	tcpClient.conn.Write(bytesData)
}

/*
func (tcpClient *TcpClient) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}

	//添加事件委托
	tcpClient.requestCallBackFuncs.Store(rpcRequest.Id, callback)

	//设定超时清理委托
	t := time.NewTimer(time.Second * time.Duration(tcpClient.timeoutInterval))
	tcpClient.requestCallBackTimers.Store(rpcRequest.Id, t)
	ctx, cancel := context.WithCancel(context.Background())
	tcpClient.requestCallBackCancels.Store(rpcRequest.Id, cancel)
	go tcpClient.TimeOutDel(ctx, rpcRequest.Id, t)

	//发送数据
	bytes := []byte{tcpClient.DATAEND}
	bytesData = append(bytesData, bytes...)
	tcpClient.conn.Write(bytesData)
}
*/

func (tcpClient *TcpClient) Push(rpcRequest *Jsonrpcrequest) {
	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}

	//发送数据
	bytes := []byte{tcpClient.DATAEND}
	bytesData = append(bytesData, bytes...)
	tcpClient.conn.Write(bytesData)
}

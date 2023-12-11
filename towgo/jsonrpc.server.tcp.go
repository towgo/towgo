/*
JSON-RPC2.0 over tcp
by:liangliangit
纯tcp socket连接
长连接、可降低http短连接带来的额外tcp握手开销，减少TIME_WAIT过多导致的端口被全部占用问题
推荐内部服务使用
*/
package towgo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
)

const (
	//DATASTART string = "[[TCP:JSONRPC:START]]"
	BUFFERLENGTH int64 = 1024000 //内存缓冲区长度 单次可以传输的最大数据量 （字节） 1024000byte=100MB
	DATAEND      byte  = '\n'    //数据尾帧标识符 (防止粘包)
)

type TcpServer struct {
	PingInterval int64
	listener     net.Listener
	cancel       context.CancelFunc
}

type TcpConn struct {
	PingInterval int64
	conn         net.Conn
	cancel       context.CancelFunc
	BUFFERLENGTH int64 //内存缓冲区长度 （字节） 1024000byte=100MB
	DATAEND      byte
}

type TcpRpcConnection struct {
	guid                   string
	isConnected            bool
	remoteAddr             string
	rpcRequest             *Jsonrpcrequest
	rpcResponse            *Jsonrpcresponse
	requestCallBackFuncs   sync.Map
	requestCallBackCancels sync.Map
	conn                   net.Conn
	bodyStr                string
	paramsBytes            []byte
	resultBytes            []byte
}

func TCPServiceHandller(tcpConn *TcpConn, data string, rpcConn *TcpRpcConnection) {
	if rpcConn == nil {
		return
	}
	rpcConn.AnalysisByString(data)
	//运行拦截器
	err := defaultJsonRpcInterceptor(rpcConn)
	if err != nil {
		return //拦截后 rpc响应由拦截器处理，  不需要再次响应
	}
	request := rpcConn.GetRpcRequest()
	if request.Method == "" {
		rpcResponse := rpcConn.GetRpcResponse()
		if rpcResponse.Id == "" {
			//无效信息
			return
		}
		//rpc响应
		f, ok := rpcConn.requestCallBackFuncs.Load(rpcResponse.Id)
		if ok {
			callback := f.(func(JsonRpcConnection))
			callback(rpcConn)
		}
		cancel, ok := rpcConn.requestCallBackCancels.Load(rpcResponse.Id)
		if ok {
			c := cancel.(context.CancelFunc)
			c()
		}
	} else { //rpc请求
		Exec(rpcConn)
	}

}

// 全双工会话
func ChatHandller(ctx context.Context, tcpConn *TcpConn) {
	rpcConn := NewTcpRpcConnection(tcpConn.conn, "{}")
	rpcConn.isConnected = true
	//读取通道
	go ConnReading(ctx, tcpConn, TCPServiceHandller, rpcConn)

	//定时ping
	for {
		time.Sleep(time.Second * time.Duration(tcpConn.PingInterval))
		select {
		case <-ctx.Done():
			return
		default:
			pingStr := "{ping}" + string(DATAEND)
			log.Print("send ping data : " + pingStr)
			tcpConn.conn.Write([]byte(pingStr))
		}
	}

}

func ConnReading(ctx context.Context, tcpConn *TcpConn, handller func(*TcpConn, string, *TcpRpcConnection), rpcConn *TcpRpcConnection) {
	readBufString := ""
	for {

		bufferData := make([]byte, BUFFERLENGTH)
		readLength, err := tcpConn.conn.Read(bufferData)
		if err != nil {
			rpcConn.isConnected = false
			if err.Error() == "EOF" {
				log.Printf("断开连接:%s \n", tcpConn.conn.RemoteAddr().String())
			} else {
				log.Printf("读取数据错误！ 错误信息：%s \n", err)
			}
			if tcpConn == nil {
				return
			}
			if tcpConn.cancel == nil {
				return
			}
			tcpConn.cancel()
			return
		}

		readBufString = string(bufferData[0:readLength])
		if strings.Contains(readBufString, string(DATAEND)) {

			strArr := strings.Split(readBufString, string(DATAEND))

			readBufString = ""
			for k, v := range strArr {
				//数据完整-可以提交给上层应用
				if k < (len(strArr) - 1) {
					go handller(tcpConn, strArr[k], rpcConn)

				} else {
					if readLength == int(BUFFERLENGTH) {
						//不完整的消息   缓存起来等待下一批数据一起组装

						readBufString = readBufString + v
						if strings.Count(readBufString, "")-1 > int(BUFFERLENGTH) {
							readBufString = ""
							log.Printf("缓冲区溢出\n")
						}
					}
				}
			}
		} else {
			if strings.Count(readBufString, "")-1 > int(BUFFERLENGTH) {
				readBufString = ""
				log.Printf("缓冲区溢出\n")
			}
		}

	}
}

func ConnReadingClient(ctx context.Context, tcpClient *TcpClient, tcpConn *TcpConn, handller func(*TcpConn, string)) {
	readBufString := ""
	for {

		bufferData := make([]byte, BUFFERLENGTH)
		readLength, err := tcpConn.conn.Read(bufferData)

		if err != nil {
			log.Printf("断开连接:%s info:%s \n", tcpConn.conn.RemoteAddr().String(), err)
			tcpClient.connected = false
			tcpClient.OnClose(tcpClient)
			if tcpConn == nil {
				return
			}
			if tcpConn.cancel == nil {
				return
			}
			tcpConn.cancel()
			return
		}

		readBufString = string(bufferData[0:readLength])
		if strings.Contains(readBufString, string(DATAEND)) {

			strArr := strings.Split(readBufString, string(DATAEND))

			readBufString = ""
			for k, v := range strArr {
				//数据完整-可以提交给上层应用
				if k < (len(strArr) - 1) {
					go handller(tcpConn, strArr[k])

				} else {
					if readLength == int(BUFFERLENGTH) {
						//不完整的消息   缓存起来等待下一批数据一起组装

						readBufString = readBufString + v
						if strings.Count(readBufString, "")-1 > int(BUFFERLENGTH) {
							readBufString = ""
							log.Printf("缓冲区溢出\n")
						}
					}
				}
			}
		} else {

			if strings.Count(readBufString, "")-1 > int(BUFFERLENGTH) {
				readBufString = ""
				log.Printf("缓冲区溢出\n")
			}

		}

	}
}

func NewTcpServer(ip, port string) (*TcpServer, error) {
	listener, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		return nil, err
	}
	tcpServer := &TcpServer{
		listener:     listener,
		PingInterval: 20,
	}
	return tcpServer, nil
}

func (tcpserver *TcpServer) Stop() {
	tcpserver.listener.Close()
	tcpserver.cancel()
}

func (tcpserver *TcpServer) Run() {
	tcpserverctx, cancel := context.WithCancel(context.Background())
	tcpserver.cancel = cancel
	go func(tcpserverctx context.Context) {
		for {
			select {
			case <-tcpserverctx.Done():
				return
			default:
				conn, e := tcpserver.listener.Accept()
				clientCtx, cancel := context.WithCancel(context.Background())
				tcpConn := &TcpConn{
					cancel:       cancel,
					conn:         conn,
					PingInterval: tcpserver.PingInterval,
					BUFFERLENGTH: BUFFERLENGTH, //内存缓冲区长度 （字节） 1024000byte=100MB
					DATAEND:      DATAEND,      //数据尾帧标识符
				}
				if e != nil {
					log.Print(e.Error())
				} else {
					go ChatHandller(clientCtx, tcpConn)
				}
			}
		}
	}(tcpserverctx)
}

// 转换为RPC连接器连接对象
func NewTcpRpcConnection(conn net.Conn, bodyStr string) *TcpRpcConnection {

	if bodyStr == "" {
		return nil
	}

	//rpc 请求
	rpcRequest, _ := ToJsonrpcrequest(bodyStr)

	rpcResponse, _ := ToJsonrpcresponse(bodyStr)

	rpcConn := &TcpRpcConnection{
		guid:        system.GetGUID().Hex(),
		rpcResponse: &rpcResponse,
		rpcRequest:  rpcRequest,
		conn:        conn,
		bodyStr:     bodyStr,
	}

	return rpcConn
}

func (c *TcpRpcConnection) WriteError(code int64, msg string) {
	c.rpcResponse.Error.Set(code, msg)
	c.Write()
}

// 获取对方ip地址
func (c *TcpRpcConnection) GetRemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

// 读取参数
func (c *TcpRpcConnection) ReadParams(destParams ...interface{}) error {
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
func (c *TcpRpcConnection) ReadResult(destParams ...interface{}) error {

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

func (c *TcpRpcConnection) Write() {
	c.rpcResponse.Id = c.rpcRequest.Id
	c.rpcResponse.Timestampin = c.rpcRequest.Timestampin
	time := time.Now().UnixNano() / 1e6
	c.rpcResponse.Timestampout = strconv.FormatInt(time, 10)
	mjson, _ := json.Marshal(c.rpcResponse)
	enddata := []byte{DATAEND}

	var buffer bytes.Buffer //Buffer是一个实现了读写方法的可变大小的字节缓冲

	buffer.Write(mjson)
	buffer.Write(enddata) //打上结束标签

	b3 := buffer.Bytes() //得到了b1+b2的结果

	c.conn.Write(b3)
}

func (c *TcpRpcConnection) WriteResult(result interface{}) {
	c.rpcResponse.Result = result
	c.Write()
}

func (c *TcpRpcConnection) WriteResponse(rpcResponse Jsonrpcresponse) {
	c.rpcResponse = &rpcResponse
	c.Write()
}

func (c *TcpRpcConnection) Read() string {

	return c.bodyStr
}

func (c *TcpRpcConnection) GetRpcRequest() *Jsonrpcrequest {
	return c.rpcRequest
}
func (c *TcpRpcConnection) GetRpcResponse() *Jsonrpcresponse {
	return c.rpcResponse
}

/*
推送请求，推送请求的设计是将请求作为一个事件发布，并且不需要对方响应。
push也可以作为异步消息使用（客户端与服务端均建立对应的method，互相push）
*/
func (c *TcpRpcConnection) Push(request *Jsonrpcrequest) error {
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
			mjson = code
		}
	}
	enddata := []byte{DATAEND}

	var buffer bytes.Buffer //Buffer是一个实现了读写方法的可变大小的字节缓冲

	buffer.Write(mjson)
	buffer.Write(enddata) //打上结束标签

	b3 := buffer.Bytes() //得到了b1+b2的结果

	_, err := c.conn.Write(b3)
	return err
}

// 连接类型
func (c *TcpRpcConnection) LinkType() string {
	return "tcp"
}

// 解析消息
func (c *TcpRpcConnection) AnalysisByString(message string) {
	//rpc 请求
	c.bodyStr = message
	rpcRequest, _ := ToJsonrpcrequest(message)
	c.rpcRequest = rpcRequest
	c.paramsBytes = nil

	if rpcRequest.Method == "" {
		response, _ := ToJsonrpcresponse(message)
		c.rpcResponse = &response
	} else {
		c.rpcResponse = NewJsonrpcresponse()
	}
}

func (c *TcpRpcConnection) Call(rpcRequest *Jsonrpcrequest, callback func(JsonRpcConnection)) {
	rpcRequest.Id = system.GetGUID().Hex()
	rpcRequest.Jsonrpc = "2.0"
	timestampin := time.Now().UnixNano() / 1e6
	rpcRequest.Timestampin = strconv.FormatInt(timestampin, 10)
	bytesData, err := json.Marshal(rpcRequest)
	if err != nil {
		return
	}
	c.requestCallBackFuncs.Store(rpcRequest.Id, callback)

	ctx, cancel := context.WithCancel(context.Background())
	c.requestCallBackCancels.Store(rpcRequest.Id, cancel)

	//delete requestCallBackFunc by requestid
	go func(ctx context.Context, requestid string) {
		t := time.NewTimer(time.Second * time.Duration(clientCallTimeOut))
		select {
		case <-ctx.Done():
			t.Stop()
		case <-t.C:
		}
		c.requestCallBackFuncs.Delete(requestid)
		c.requestCallBackCancels.Delete(requestid)
	}(ctx, rpcRequest.Id)
	c.conn.Write(bytesData)
}

func (c *TcpRpcConnection) IsConnected() bool {
	return c.isConnected
}

func (c *TcpRpcConnection) GUID() string {
	return c.guid
}

func (c *TcpRpcConnection) Close() {
	c.conn.Close()
}

func (c *TcpRpcConnection) EnableHealthCheck() {
}

func (c *TcpRpcConnection) DisableHealthCheck() {
}

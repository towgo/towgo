/*
代码暂未实现功能     开发中...
*/

package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/towgo"
)

const (
	END_DATA = '\n'
)

type ConnectionPipe struct {
	callTimeoutInterval    int64
	receiveTimeoutInterval int64
	ioLocker               sync.Mutex
	rpcResponses           sync.Map
	baseConn               io.ReadWriteCloser
	rpcConn                towgo.JsonRpcConnection
	maxBufLen              int64
	dataBuf                []byte
}

// 新建jsonrcp管道  实现数据透传
func NewJsonrpcPipe(baseConn io.ReadWriteCloser, rpcConn towgo.JsonRpcConnection) *ConnectionPipe {
	newConn := &ConnectionPipe{
		baseConn:               baseConn,
		rpcConn:                rpcConn,
		callTimeoutInterval:    60,
		receiveTimeoutInterval: 10,
	}
	go newConn.baseConnHandller()
	return newConn
}

// 开始监听原始线路
func (c *ConnectionPipe) baseConnHandller() {
	for {
		buf := make([]byte, 1024)
		n, err := c.baseConn.Read(buf)
		if err != nil {
			log.Print(err.Error())
			return
		}
		c.ioLocker.Lock()
		if len(c.dataBuf) > int(c.maxBufLen) {
			log.Print("缓冲溢出,默认将数据丢弃")
		} else {
			c.dataBuf = append(c.dataBuf, buf[0:n]...)
		}

		//尝试解析
		if bytes.Contains(c.dataBuf, []byte{END_DATA}) {
			//尾帧存在，继续解析
			rpcByte, nextByte := SplitBytes(c.dataBuf, []byte{END_DATA})
			resp := towgo.NewJsonrpcresponse()
			err := json.Unmarshal(rpcByte, resp)
			if err != nil {
				log.Print(err.Error())
			} else {
				c.rpcResponses.Store(resp.Id, resp)
				go func(c *ConnectionPipe, id string) {
					time.Sleep(time.Second * time.Duration(c.receiveTimeoutInterval))
					c.rpcResponses.Delete(id)
				}(c, resp.Id)
			}
			c.dataBuf = nextByte
		}
		c.ioLocker.Unlock()
	}
}

// 一次推送
func (c *ConnectionPipe) Push(method string, params any) error {
	//组装数据
	request := EncodeConnectionRPCData(method, params)
	request.Id = system.GetGUID().Hex()
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	//补上尾帧防止粘包
	b = append(b, END_DATA)

	//发送数据
	_, err = c.conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

// 一次请求
func (c *ConnectionPipe) Call(method string, params any, destResult any) error {
	//组装数据
	request := EncodeConnectionRPCData(method, params)
	request.Id = system.GetGUID().Hex()
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}

	//补上尾帧防止粘包
	b = append(b, END_DATA)

	//发送数据
	_, err = c.conn.Write(b)
	if err != nil {
		return err
	}
	result, err := c.receiveJsonrpc(request.Id)
	if err != nil {
		return err
	}

	if destResult != nil {
		err = result.ReadResult(destResult)
	}
	return err
}

func (c *ConnectionPipe) receiveJsonrpc(requestID string) (*towgo.Jsonrpcresponse, error) {
	timer := time.NewTimer(time.Second * time.Duration(c.callTimeoutInterval))
	for {
		select {
		case <-timer.C:
			return nil, errors.New("请求超时")
		default:
			resp, ok := c.rpcResponses.LoadAndDelete(requestID)
			if ok {
				resp.(*towgo.Jsonrpcresponse)
			}
			continue
		}
	}
}

func EncodeConnectionRPCData(method string, params any) *towgo.Jsonrpcrequest {
	request := towgo.NewJsonrpcrequest()
	request.Method = method
	request.Params = params
	return request
}

// splitBytes 根据特殊字节切片的位置，将输入字节切片分割为两个部分
func SplitBytes(input []byte, special []byte) ([]byte, []byte) {
	// 查找特殊字节切片在输入切片中的位置
	startIndex := bytes.Index(input, special)
	if startIndex == -1 {
		// 如果未找到特殊字节切片，则返回整个输入切片和空切片
		return input, nil
	}

	// 计算特殊字节切片的结束位置
	endIndex := startIndex + len(special)

	// 分割输入切片
	beforeSpecial := input[:startIndex]
	afterSpecial := input[endIndex:]
	return beforeSpecial, afterSpecial
}

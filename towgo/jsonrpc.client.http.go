/*
json rpc 2.0 http客户端
by:liangliangit
*/
package towgo

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/towgo/towgo/lib/system"
)

type HttpClient struct {
	ErrorFunc func(error)
}

func NewHttpClient() *HttpClient {
	HttpClient := &HttpClient{
		ErrorFunc: func(err error) {
			log.Print(err)
		},
	}
	return HttpClient
}

func (c *HttpClient) Call(url string, jsonrpcrequest *Jsonrpcrequest, callback func(Jsonrpcresponse)) {

	if jsonrpcrequest.Id == "" {
		jsonrpcrequest.Id = system.RandChar(64)
	}

	if jsonrpcrequest.Timestampin == "" {
		time := time.Now().UnixNano() / 1e6
		jsonrpcrequest.Timestampin = strconv.FormatInt(time, 10)
	}

	bytesData, err := json.Marshal(jsonrpcrequest)

	if isencryption {
		strData, err := AesEncrypt(bytesData)

		if err != nil {
			log.Print(err.Error())
			return
		}
		bytesData = []byte(strData)
	}

	if err != nil {
		c.ErrorFunc(err)
		return
	}
	reader := bytes.NewReader(bytesData)

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		c.ErrorFunc(err)
		return
	}

	c.post(jsonrpcrequest, request, callback)

}

func (c *HttpClient) Push(url string, jsonrpcrequest *Jsonrpcrequest) {

	if jsonrpcrequest.Id == "" {
		jsonrpcrequest.Id = system.RandChar(64)
	}

	if jsonrpcrequest.Timestampin == "" {
		time := time.Now().UnixNano() / 1e6
		jsonrpcrequest.Timestampin = strconv.FormatInt(time, 10)
	}

	bytesData, err := json.Marshal(jsonrpcrequest)

	if isencryption {
		strData, err := AesEncrypt(bytesData)

		if err != nil {
			log.Print(err.Error())
			return
		}
		bytesData = []byte(strData)
	}

	if err != nil {
		c.ErrorFunc(err)
		return
	}
	reader := bytes.NewReader(bytesData)

	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		c.ErrorFunc(err)
		return
	}

	c.post(jsonrpcrequest, request, nil)

}

// 发起post请求
func (c *HttpClient) post(HttpClientRequest *Jsonrpcrequest, request *http.Request, callback func(Jsonrpcresponse)) {
	//程序退出后删除委托
	//defer c.deleteFunc(HttpClientRequest.Id)

	HttpClient := http.Client{}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	resp, err := HttpClient.Do(request)
	if err != nil {
		c.ErrorFunc(err)
		return
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		c.ErrorFunc(err)
		return
	}

	bodyStr := string(respBytes)
	jsonrpcresponse, err := ToJsonrpcresponse(bodyStr)
	if err != nil {
		c.ErrorFunc(err)
		return
	}

	//委托运行函数
	if callback != nil {
		callback(jsonrpcresponse)
	}
}

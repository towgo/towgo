package iolog

import (
	"errors"
	"log"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
)

const (
	STATUS_QUEUE              = "STATUS_QUEUE"              //队列中
	STATUS_WAIT_BACK_RESPONSE = "STATUS_WAIT_BACK_RESPONSE" //等待后端响应
	STATUS_WAIT_BACK_TIMEOUT  = "STATUS_WAIT_BACK_TIMEOUT"  //后端响应超时
	STATUS_PROXY_ERROR        = "STATUS_PROXY_ERROR"        //代理异常
	STATUS_CLIENT_CANCEL      = "STATUS_CLIENT_CANCEL"      //客户端取消
	STATUS_DONE               = "STATUS_DONE"               //完成
)

func (IoLog) TableName() string {
	return "io_log"
}

type IoLog struct {
	Guid           string `json:"guid" xorm:"pk"`
	Uid            int64  `json:"uid"`
	Username       string `json:"username"`
	RemoteHost     string `json:"remote_host"`
	RequestUrl     string `json:"request_url"`
	RequestApiPath string `json:"request_api_path"`
	ProxyHost      string `json:"proxy_host"`
	Result         string `json:"result"`
	RequestTime    int64  `json:"request_time"`
	ProxyTime      int64  `json:"proxy_time"`
	Status         string `json:"status"`
	ResponseTime   int64  `json:"response_time"`
}

var LOG_EXPIRE_TIME int64 = 86400 * 180 //日志保存180天

func init() {
	xormDriver.Sync2(new(IoLog))
	autoClear()
}

func autoClear() {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Print(err)
			}
			autoClear()
		}()
		var model IoLog
		for {
			time.Sleep(time.Second * 60)
			befor := time.Now().Unix() - LOG_EXPIRE_TIME
			sql := "delete from " + model.TableName() + " where request_time < ?"
			basedboperat.SqlExec(sql, befor)
		}
	}()
}

func WriteRequest(Guid string, Uid int64, RemoteHost, RequestUrl, RequestApiPath, ProxyHost string) (err error) {
	i := IoLog{}
	i.Guid = Guid
	i.Uid = Uid
	i.RemoteHost = RemoteHost
	i.RequestUrl = RequestUrl
	i.RequestApiPath = RequestApiPath
	i.ProxyHost = ProxyHost
	i.RequestTime = time.Now().Unix()
	_, err = basedboperat.Create(&i)
	return
}

func WriteResult(guid string, Result string) error {
	if guid == "" {
		return errors.New("id can not be null")
	}
	i := IoLog{}
	i.Result = Result
	i.ResponseTime = time.Now().Unix()
	i.Status = STATUS_DONE
	return basedboperat.Update(i, []string{"result", "response_time", "status"}, "guid = ?", guid)
}

func SetLogExpireTime(t int64) {
	LOG_EXPIRE_TIME = t
}

/*
日志模块
*/
package dblog

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/towgo/towgo/dao/basedboperat"
)

// 日志最长保留时间 单位秒
var maxSaveTime int64 = 86400 * 30

func (Log) TableName() string {
	return "logs"
}

type Log struct {
	ID            int64  `json:"id"`                                                        //id
	ServiceName   string `json:"service_name" xorm:"varchar(125)" gorm:"type:varchar(256)"` //服务名
	ServiceIp     string `json:"service_ip" xorm:"varchar(125)" gorm:"type:varchar(256)"`   //服务ip
	ModuleName    string `json:"module_name" xorm:"varchar(125)" gorm:"type:varchar(256)"`  //模块名
	Protocol      string `json:"protocol" xorm:"varchar(125)" gorm:"type:varchar(256)"`     //协议
	RequestParam  string `json:"request_param"`                                             //入参
	ResponseParam string `json:"response_param"`                                            //出参
	RequestIp     string `json:"request_ip" xorm:"varchar(125)" gorm:"type:varchar(256)"`   //请求
	Method        string `json:"method" xorm:"varchar(256)" gorm:"type:varchar(256)"`       //请求路径
	CreatedAt     int64  `json:"created_at" xorm:"created"`                                 //创建时间
	UpdatedAt     int64  `json:"updated_at" xorm:"updated"`                                 //更新时间
	CreatedBy     string `json:"created_by" xorm:"varchar(125)" gorm:"type:varchar(256)"`   //创建人
	Status        int64  `json:"status"`
	Detail        string `json:"detail" xorm:"longtext"`
	OperateType   int64  `json:"operate_type" xorm:"-"`
	MethodName    string `json:"method_name" xorm:"-"`
}

func (l *Log) AfterQuery() {
	var ot OperateType
	basedboperat.Get(&ot, nil, "method =? and service_name =? and module_name = ? ", l.Method, l.ServiceName, l.ModuleName)
	l.OperateType = ot.Type
	l.MethodName = ot.MethodName

}
func clearExp() {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Print(err)
			}
			clearExp()
		}()
		for {
			time.Sleep(time.Second * 60)
			befor := time.Now().Unix() - maxSaveTime
			sql := "delete from logs where created_at < ?"
			basedboperat.SqlExec(sql, befor)
		}
	}()
}

// 将日志保存到数据库
func (l *Log) Save() error {
	l.ID = 0 //置零ID用于数据库自动生成流水号ID
	_, err := basedboperat.Create(l)
	return err
}

// 写入日志信息
func Write(serviceName, serviceIp, moduleName, protocol, requestParam, responseParam, requestIp, method, createBy string, status int64, detail string) error {
	go func() {
		log := Log{
			ServiceName:   serviceName,
			ServiceIp:     serviceIp,
			ModuleName:    moduleName,
			Protocol:      protocol,
			RequestParam:  requestParam,
			ResponseParam: responseParam,
			RequestIp:     requestIp,
			Method:        method,
			CreatedBy:     createBy,
			Status:        status,
			Detail:        detail,
		}
		log.Save()
	}()
	return nil
}

//	func UpdateWrite(uuid, responseParam string) error {
//		go func() {
//			log := Log{
//				ResponseParam: responseParam,
//			}
//			basedboperat.Update(&log,nil,"uuid = ?",uuid)
//		}()
//		return nil
//	}
type DownloadLogFileReq struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

func UnixToLoaclTime(t int64) string {
	return time.Unix(t/1000, 0).Format(time.DateTime)
}

func GetUuid() string {
	u4 := uuid.New()
	uuid := strings.Split(u4.String(), "-")[4]
	return uuid
}

func PathIfNotExistsCreate(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, err
}

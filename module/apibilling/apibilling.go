package apibilling

/*
用户账单计费模块
by:liangliangit
*/
import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/dlocker"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/system"
	"github.com/towgo/towgo/module/accountcenter"
)

const (
	UNLIMIT = -1
)

func init() {
	xormDriver.Sync2(new(ApiBilling))
}

func (ApiBilling) TableName() string {
	return "api_billing"
}

type ApiBilling struct {
	ID                        string   `json:"id" xorm:"pk"`
	Uid                       int64    `json:"uid"` //关联用户唯一标识
	UserNickname              string   `json:"user_nickname"`
	ApiPath                   string   `json:"api_path"`                                        //路径
	CountLimit                int64    `json:"count_limit"`                                     //次数限制
	NotifyMails               []string `json:"notify_mails" xorm:"json" gorm:"serializer:json"` //通知邮箱
	NotifyBeforeExpireTimeDay int64    `json:"notify_before_expire_time_day"`                   //到期前的N天进行通知
	TotalRequestCount         int64    `json:"total_request_count"`                             //总请求次数
	TotalFailedCount          int64    `json:"total_failed_count"`                              //总失败数
	CountPrice                float64  `json:"count_price"`                                     //每次消费金额
	ExpireTime                int64    `json:"expire_time"`                                     //到期时间
}

// 计费
func Charging(uid int64, apiPath string) error {
	ab := Matching(uid, apiPath)
	if ab != nil {
		return ab.Charging()
	}
	return nil
}

// 退款
func Refund(uid int64, apiPath string) error {
	ab := Matching(uid, apiPath)
	if ab != nil {
		return ab.Refund()
	}
	return nil
}

func Matching(uid int64, apiPath string) *ApiBilling {
	var ab ApiBilling
	var abs []ApiBilling
	var list basedboperat.List
	list.And = map[string][]interface{}{
		"uid": []interface{}{uid},
	}
	list.Limit = -1
	basedboperat.ListScan(&list, ab, &abs)
	for _, v := range abs {
		if v.IsMatching(apiPath) {
			return &v
		}
	}
	return nil
}

// 检查路径是否可计费
func (a *ApiBilling) IsMatching(apiPath string) bool {
	//模糊匹配
	paths := strings.Split(apiPath, "/")
	var mathPaths []string
	var path string
	pathsLen := len(paths)

	mathPaths = append(mathPaths, "/")
	for i := 0; i < pathsLen; i++ {
		if paths[i] != "" {
			path = path + "/" + paths[i]
		}
		mathPaths = append(mathPaths, path)
	}

	for i := len(mathPaths) - 1; i >= 0; i-- {
		if a.ApiPath == mathPaths[i] {
			return true
		}
	}
	return false
}

func (a *ApiBilling) IsCanUse() bool {
	return (a.CountLimit > 0 && time.Now().Unix() < a.ExpireTime)
}

func (a *ApiBilling) GenMethod() string {
	return a.ID
}

// 单次计费
func (a *ApiBilling) Charging() error {
	dlocker.Lock(a.GenMethod(), 1)
	defer dlocker.UnLock(a.GenMethod())
	if a.CountLimit == UNLIMIT {
		return basedboperat.SqlExec("update "+a.TableName()+" set total_request_count = total_request_count + 1 where id = ?", a.ID)
	}
	if !a.IsCanUse() {
		return errors.New("API无法使用! 次数已使用完毕或已经到期")
	}
	return basedboperat.SqlExec("update "+a.TableName()+" set count_limit = count_limit - 1,total_request_count = total_request_count + 1 where id = ? and count_limit > 0", a.ID)
}

// 单次退款
func (a *ApiBilling) Refund() error {
	dlocker.Lock(a.GenMethod(), 1)
	defer dlocker.UnLock(a.GenMethod())

	//无限制不需要退款
	if a.CountLimit == UNLIMIT {
		return basedboperat.SqlExec("update "+a.TableName()+" set total_failed_count = total_failed_count + 1 where id = ?", a.ID)
	}
	//计算退款
	return basedboperat.SqlExec("update "+a.TableName()+" set count_limit = count_limit + 1,total_failed_count = total_failed_count + 1 where id = ? and count_limit > 0", a.ID)
}

func (a *ApiBilling) BeforCreate(dbSession basedboperat.DbTransactionSession) error {

	session := dbSession.Value(jsonrpc.SESSION)
	var requestParams struct {
		ID int64 `json:"id"`
	}
	requestParams.ID = a.Uid
	var account accountcenter.Account
	err := jsonrpc.CallEdgeServerNode("/account/query", session.(string), requestParams, &account)
	if err != nil {
		return err
	}
	if account.ID == 0 {
		return errors.New("用户不存在")
	}
	a.UserNickname = account.Nickname

	a.ID = system.MD5(strconv.FormatInt(a.Uid, 10) + a.ApiPath)

	return nil
}

func (a *ApiBilling) BeforUpdate(dbSession basedboperat.DbTransactionSession) error {
	oldId := a.ID
	newId := system.MD5(strconv.FormatInt(a.Uid, 10) + a.ApiPath)

	var findModel ApiBilling

	dbSession.Get(&findModel, nil, "id = ?", oldId)
	if findModel.ID == "" {
		return errors.New("记录不存在")
	}

	if newId != oldId {
		return errors.New("无法修改所属用户和计费路径")
	}
	return nil
}

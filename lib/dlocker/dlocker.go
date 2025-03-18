/*
分布式锁
基于mysql数据库
*/
package dlocker

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/system"
)

var lockersLock sync.Mutex
var lockers sync.Map
var globalCheckRate int64 = 1
var lockExpireLimit int64 = 120

func init() {
	xormDriver.Sync2(new(Locker))
	autoClear()
}

func SetCheckRate(rate int64) {
	globalCheckRate = rate
}

func autoClear() {
	go func() {
		var l Locker
		for {
			time.Sleep(time.Second * 10)
			basedboperat.SqlExec("delete from "+l.TableName()+" where created_at < ?", time.Now().Unix()-lockExpireLimit)
		}
	}()
}

func (Locker) TableName() string {
	return "locker"
}

type Locker struct {
	sync.Mutex `json:"-" xorm:"-" gorm:"-"`
	Guid       string `json:"guid" xorm:"unique"`
	Method     string `json:"method" xorm:"unique"`
	CreatedAt  int64  `json:"created_at" xorm:"created"`
}

func Lock(method string, checkRate int64) (guid string, err error) {
	if method == "" {
		return "", errors.New("method can not be null")
	}
	if checkRate <= 0 {
		checkRate = globalCheckRate
	}
	var locker *Locker
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	lockersLock.Unlock()
	if ok {
		locker = lockerAny.(*Locker)
		locker.CreatedAt = time.Now().Unix()
		locker.Lock()
	} else {
		locker = &Locker{
			Guid:   system.GetGUID().Hex(),
			Method: method,
		}
		locker.Lock()
		lockersLock.Lock()
		lockers.Store(method, locker)
		lockersLock.Unlock()
	}

	for {
		_, err := basedboperat.Create(locker)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(checkRate))
			continue
		} else {
			break
		}
	}
	return locker.Guid, nil
}

func UnLock(method string) error {
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	lockersLock.Unlock()
	var locker *Locker
	if ok {
		locker = lockerAny.(*Locker)
		basedboperat.Delete(locker, nil, "guid = ? and method = ?", locker.Guid, locker.Method)
		locker.Unlock()
	}
	return nil
}

func Renewal(guid string) {
	var locker Locker
	basedboperat.Get(&locker, nil, "guid = ?", guid)
	if locker.Method == "" {
		log.Print("lock method not found")
		return
	}
	basedboperat.SqlExec("update "+locker.TableName()+" set created_at = ? where guid = ?", time.Now().Unix(), locker.Guid)
}

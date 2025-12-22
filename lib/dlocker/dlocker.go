/*
分布式锁
基于mysql数据库
支持自定义锁过期时间 + 精准续期
*/
package dlocker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/system"
)

var (
	lockersLock     sync.Mutex
	lockers         sync.Map
	globalCheckRate int64           = 1   // 锁争抢轮询间隔（毫秒）默认值
	globalExpireSec int64           = 120 // 锁默认过期秒数（替代原lockExpireLimit）
	cleanTicker     *time.Ticker          // 自动清理定时器（便于外部控制）
	cleanCtx        context.Context       // 清理协程上下文（优雅退出）
	cleanCancel     context.CancelFunc
)

func init() {
	// 1. 同步表结构（包含新增的ExpireSeconds字段）
	basedboperat.Sync(new(Locker))
	// 2. 初始化清理协程上下文
	cleanCtx, cleanCancel = context.WithCancel(context.Background())

	// 3. 启动自动清理协程
	autoClear()
}

// SetCheckRate 设置锁争抢的默认轮询间隔（毫秒）
func SetCheckRate(rate int64) {
	if rate > 0 {
		globalCheckRate = rate
	}
}

// SetGlobalExpireSec 设置锁的全局默认过期秒数（替代原lockExpireLimit）
func SetGlobalExpireSec(sec int64) {
	if sec > 0 {
		globalExpireSec = sec
	}
}

// Locker 锁结构体（新增ExpireSeconds字段）
type Locker struct {
	sync.Mutex    `json:"-" xorm:"-" gorm:"-"`
	Guid          string `json:"guid" xorm:"unique"`                         // 锁唯一标识
	Method        string `json:"method" xorm:"unique"`                       // 锁对应的方法/资源名
	CreatedAt     int64  `json:"created_at" xorm:"created"`                  // 创建时间戳（秒）
	ExpireSeconds int64  `json:"expire_seconds" xorm:"not null default 120"` // 单个锁的过期秒数
}

// TableName 指定数据库表名
func (Locker) TableName() string {
	return "locker"
}

// autoClear 自动清理过期锁（按单个锁的ExpireSeconds清理）
func autoClear() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					log.Println("err:", err)
				}
			}
		}()

		// 10秒执行一次清理（可根据业务调整）
		cleanTicker = time.NewTicker(10 * time.Second)
		defer cleanTicker.Stop()

		var l Locker
		for {
			select {
			case <-cleanCtx.Done():

				return
			case <-cleanTicker.C:
				// 清理条件：created_at + expire_seconds < 当前时间戳（秒）
				now := time.Now().Unix()
				sql := "DELETE FROM " + l.TableName() + " WHERE (created_at + expire_seconds) < ?"
				basedboperat.SqlExec(sql, now)

			}
		}
	}()
}

// Lock 加锁（阻塞式，直到获取锁）
// 参数：
//
//	method: 锁对应的方法/资源名（不能为空）
//	checkRate: 争抢锁的轮询间隔（毫秒，<=0则使用全局默认值）
//	expireSec: 锁过期秒数（<=0则使用全局默认值120秒）
//
// 返回：
//
//	guid: 锁唯一标识（用于续期/解锁）
//	err: 错误信息
func Lock(method string, checkRate int64, expireSec ...int64) (guid string, err error) {
	// 1. 参数校验
	if method == "" {
		return "", errors.New("method can not be null")
	}
	// 2. 轮询间隔（毫秒）
	if checkRate <= 0 {
		checkRate = globalCheckRate
	}
	// 3. 过期秒数（默认全局120秒）
	expire := globalExpireSec
	if len(expireSec) > 0 && expireSec[0] > 0 {
		expire = expireSec[0]
	}

	// 4. 初始化/获取本地锁对象
	var locker *Locker
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	if ok {
		locker = lockerAny.(*Locker)
		locker.CreatedAt = time.Now().Unix() // 更新创建时间
		locker.ExpireSeconds = expire        // 更新过期时间
	} else {
		locker = &Locker{
			Guid:          system.GetGUID().Hex(),
			Method:        method,
			ExpireSeconds: expire,
		}
		lockers.Store(method, locker)
	}
	lockersLock.Unlock()

	locker.Lock()         // 加本地互斥锁，防止并发争抢
	defer locker.Unlock() // 最终解锁

	// 5. 阻塞式争抢锁（直到成功）
	for {
		_, err := basedboperat.Create(locker)
		if err == nil {
			return locker.Guid, nil
		}
		// 争抢失败，休眠后重试
		time.Sleep(time.Millisecond * time.Duration(checkRate))
	}
}

// TryLock 尝试加锁（非阻塞式，一次尝试）
// 参数同Lock
func TryLock(method string, expireSec ...int64) (guid string, err error) {
	// 1. 参数校验
	if method == "" {
		return "", errors.New("method can not be null")
	}

	// 2. 过期秒数
	expire := globalExpireSec
	if len(expireSec) > 0 && expireSec[0] > 0 {
		expire = expireSec[0]
	}

	// 3. 初始化/获取本地锁对象
	var locker *Locker
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	if ok {
		locker = lockerAny.(*Locker)
		locker.CreatedAt = time.Now().Unix()
		locker.ExpireSeconds = expire
	} else {
		locker = &Locker{
			Guid:          system.GetGUID().Hex(),
			Method:        method,
			ExpireSeconds: expire,
		}
		lockers.Store(method, locker)
	}
	lockersLock.Unlock()

	locker.Lock()
	defer locker.Unlock()

	// 4. 单次尝试创建锁
	_, err = basedboperat.Create(locker)
	if err != nil {
		return "", err
	}
	return locker.Guid, nil
}

// UnLock 解锁
func UnLock(guid, method string) error {
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	lockersLock.Unlock()

	if !ok {
		err := errors.New("lock not found: " + method)
		return err
	}

	locker := lockerAny.(*Locker)
	locker.Lock()
	defer locker.Unlock()

	// 删除数据库中的锁记录
	i, err := basedboperat.Delete(locker, nil, "guid = ? and method = ?", guid, locker.Method)
	if err != nil {

		return err
	}
	if i == 0 {
		return fmt.Errorf(" failed to unlock locker with method: %s update rows %d", locker.Method, i)
	}

	// 移除本地缓存的锁
	lockers.Delete(method)
	return nil
}

// Renewal 锁续期（精准续期，延长过期时间）
// 参数：
//
//	guid: 锁的唯一标识（Lock/TryLock返回的guid）
//	addExpireSec: 新增续期秒数（<=0则沿用锁原有过期秒数）
//
// 返回：
//
//	err: 续期失败原因
func Renewal(guid string, addExpireSec ...int64) error {
	if guid == "" {
		return errors.New("guid can not be null")
	}

	// 1. 查询锁信息
	var locker Locker
	err := basedboperat.Get(&locker, nil, "guid = ?", guid)
	if err != nil {

		return err
	}
	if locker.Method == "" {
		err := errors.New("lock not found by guid: " + guid)

		return err
	}

	// 2. 确定续期秒数（默认沿用原有过期秒数）
	renewSec := locker.ExpireSeconds
	if len(addExpireSec) > 0 && addExpireSec[0] > 0 {
		renewSec = addExpireSec[0]
	}

	// 3. 加本地锁，防止并发续期
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(locker.Method)
	lockersLock.Unlock()
	if !ok {
		err := errors.New("local lock not found: " + locker.Method)

		return err
	}
	localLocker := lockerAny.(*Locker)
	localLocker.Lock()
	defer localLocker.Unlock()

	// 4. 更新数据库中锁的创建时间（续期核心：created_at = 当前时间 → 过期时间顺延）
	now := time.Now().Unix()
	sql := "UPDATE " + locker.TableName() + " SET created_at = ?, expire_seconds = ? WHERE guid = ?"
	err = basedboperat.SqlExec(sql, now, renewSec, guid)
	if err != nil {

		return err
	}

	// 5. 更新本地锁信息
	localLocker.CreatedAt = now
	localLocker.ExpireSeconds = renewSec

	return nil
}

// （原有代码：lockersLock、lockers、Locker 结构体等不变）

// RenewalByKey 基于锁的Key（method）续期（更易用，入参为time.Duration）
// 参数：
//
//	method: 锁的唯一Key（对应Lock/TryLock的method参数）
//	duration: 续期时长（自动转为秒，<=0则沿用锁原有过期秒数）
//
// 返回：
//
//	err: 续期失败原因
func RenewalByKey(method string, duration time.Duration) error {
	// 1. 基础参数校验
	if method == "" {
		return errors.New("method (lock key) can not be null")
	}

	// 2. 将time.Duration转为秒（向下取整，如5*time.Minute → 300秒）
	var renewSec int64
	if duration > 0 {
		renewSec = int64(duration.Seconds())
	}
	// 3. 从本地缓存获取锁对象（加锁防并发）
	lockersLock.Lock()
	lockerAny, ok := lockers.Load(method)
	lockersLock.Unlock()

	// 4. 检查本地锁是否存在
	if !ok {
		err := errors.New("local lock not found by method: " + method)
		return err
	}
	localLocker := lockerAny.(*Locker)

	// 5. 检查锁的GUID是否有效
	if localLocker.Guid == "" {
		err := errors.New("lock guid is empty for method: " + method)
		return err
	}

	// 6. 复用原有Renewal逻辑（传入guid和续期秒数）
	// 若renewSec<=0，Renewal内部会沿用锁原有过期秒数
	return Renewal(localLocker.Guid, renewSec)
}

// StopClean 停止自动清理协程（优雅退出）
func StopClean() {
	cleanCancel()
	if cleanTicker != nil {
		cleanTicker.Stop()
	}

}

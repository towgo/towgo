package accountcenter

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

// 缓存有效期
var memCacheTimer int64 = 60 * 10

// token有效期  秒单位计算
var expirationLimit int64 = 86400 * 20

var autoClearLimit int64 = 60 * 10 //10分钟清理一次过期的token

// var expirationLimit int64 = 60
var memCache *MemCache

// UserToken结构体
func (UserToken) TableName() string {
	return "accounts_token"
}

type UserToken struct {
	TokenKey   string `gorm:"primary" xorm:"pk"`
	Uid        int64
	Payload    any `gorm:"-" xorm:"-"`
	Expiration int64
	UpdatedAt  int64
	CreatedAt  int64
}

func InitTokenTask() {
	memCache = &MemCache{}
	autoTimeToClear()
}

type MemCache struct {
	timers      sync.Map
	cacheObject sync.Map
	cancels     sync.Map
}

func (mc *MemCache) CreateTimerToDelete(tokenKey string) {
	timer := time.NewTimer(time.Second * time.Duration(memCacheTimer))
	mc.timers.Store(tokenKey, timer)
	ctx, cancel := context.WithCancel(context.Background())

	mc.cancels.Store(tokenKey, cancel)
	go mc.DeleteWhenTimeOut(ctx, tokenKey, timer)
}

func (mc *MemCache) DeleteWhenTimeOut(ctx context.Context, tokenKey string, timer *time.Timer) {
	select {
	case <-timer.C:
		mc.timers.Delete(tokenKey)
		mc.cacheObject.Delete(tokenKey)
		mc.cancels.Delete(tokenKey)
		return
	case <-ctx.Done():
		return
	}
}

func (mc *MemCache) ResetTimer(tokenKey string) {
	timerInterface, ok := mc.timers.Load(tokenKey)
	if ok {
		timer := timerInterface.(*time.Timer)
		timer.Reset(time.Second * time.Duration(memCacheTimer))
	}
}

func (mc *MemCache) Add(tokenKey string, value any) {
	mc.cacheObject.Store(tokenKey, value)
	mc.CreateTimerToDelete(tokenKey)
}

func (mc *MemCache) Del(tokenKey string) {
	timerInterface, ok := mc.timers.Load(tokenKey)
	if ok {
		timer := timerInterface.(*time.Timer)
		timer.Stop()                                               //关闭定时器
		cancel_any, isLoaded := mc.cancels.LoadAndDelete(tokenKey) //关闭定时器线程
		if isLoaded {
			cancel := cancel_any.(context.CancelFunc)
			if cancel != nil {
				cancel()
			}
		}
		mc.timers.Delete(tokenKey) //清除定时器委托
	}
	mc.cacheObject.Delete(tokenKey) //清除内存
}

func (mc *MemCache) Get(tokenKey string) (any, bool) {
	if mc == nil {
		return nil, false
	}
	return mc.cacheObject.Load(tokenKey)
}

// 自动清理过期token定时器
func autoTimeToClear() {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Print(err)
			}
			autoTimeToClear()
		}()
		var userToken UserToken
		for {
			time.Sleep(time.Second * time.Duration(autoClearLimit))
			basedboperat.SqlExec("delete from "+userToken.TableName()+" where expiration < ?", time.Now().Unix())
		}
	}()
}

func (t *UserToken) NewTokenGUID(salt string) {
	guid := system.GetGUID().Hex()
	saltEncode := system.SHA1(salt)
	tokenCode := system.SHA1(guid + saltEncode)
	t.TokenKey = tokenCode
}

// 返回一个唯一标识的token令牌
func (t *UserToken) String() string {
	return t.TokenKey
}

// token是否有效 检查有效期
// 有效返回true
// 无效返回false
func (t *UserToken) Valid() bool {
	return time.Now().Unix() < t.Expiration
}

func (t *UserToken) Check(tokenKey string) bool {
	return tokenKey == t.TokenKey
}

// 更新token有效期
func (t *UserToken) Update(expiration int64) {
	if expiration > 0 {
		t.Expiration = time.Now().Unix() + expiration
	} else {
		t.Expiration = time.Now().Unix() + expirationLimit
	}
}

// 新建一个token
func NewToken(user *Account) *UserToken {
	timenow := time.Now().Unix()
	token := &UserToken{
		CreatedAt:  timenow,
		Uid:        user.ID,
		Payload:    user,
		Expiration: timenow + expirationLimit,
	}
	user.Token = token.TokenKey
	token.NewTokenGUID(user.Salt)
	basedboperat.Create(token)
	memCache.Add(token.TokenKey, token)
	return token
}

func GetToken(tokenKey string) (*UserToken, error) {
	var userToken *UserToken = &UserToken{}

	//缓存查询
	userTokenInterface, ok := memCache.Get(tokenKey)
	if ok {
		//缓存命中
		userToken = userTokenInterface.(*UserToken)
		//缓存过期  清理
		if !userToken.Valid() {
			memCache.Del(tokenKey)
			log.Println("hit cache token过期(登录失效,请重新登录) token =" + tokenKey)
			return nil, errors.New("hit cache token过期(登录失效,请重新登录) token =" + tokenKey)
		}
	} else {
		//数据库查询

		//查询持久化数据
		err := basedboperat.Get(userToken, nil, "token_key = ?", tokenKey)
		if err != nil {
			return nil, err
		}
		if userToken.Uid == 0 {
			log.Println("select database  token不存在(登录失效,请重新登录) token =" + tokenKey)
			return nil, errors.New("select database  token不存在(登录失效,请重新登录) token =" + tokenKey) //数据不存在
		}

		//token过期
		if !userToken.Valid() {
			log.Println("select database  token不存在(登录失效,请重新登录) token =" + tokenKey)

			return nil, errors.New("select database  token过期(登录失效,请重新登录) token=" + tokenKey)
		}

		//查询token关联的用户
		var user *Account = &Account{}
		basedboperat.Get(user, nil, "id = ?", userToken.Uid)
		if user.ID == 0 {
			log.Println(fmt.Sprintf("select database  token关联用户不存在(登录失效,请重新登录) userToken.Uid =%v", userToken.Uid))
			return nil, errors.New(fmt.Sprintf("select database  token关联用户不存在(登录失效,请重新登录) userToken.Uid =%v", userToken.Uid)) //用户不存在
		}

		//写入缓存
		userToken.Payload = user
		memCache.Add(userToken.TokenKey, userToken)
	}

	return userToken, nil
}

func DeleteToken(tokenKey string) {
	memCache.Del(tokenKey)
	var userToken *UserToken = &UserToken{}
	userToken.TokenKey = tokenKey

	basedboperat.Delete(userToken, nil, "token_key = ?", tokenKey)

}

func SetExpiration(hour int64) {
	expirationLimit = 3600 * hour
}

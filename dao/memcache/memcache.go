package memcache

import (
	"log"
	"sync"
	"time"
)

var memCache *MemCache

func init() {
	memCache = &MemCache{}
}

type MemCache struct {
	timers      sync.Map
	cacheObject sync.Map
}

func (mc *MemCache) CreateTimerToDelete(key string, expirationLimit int64) {
	log.Print(expirationLimit)
	timer := time.NewTimer(time.Second * time.Duration(expirationLimit))
	mc.timers.Store(key, timer)
	go mc.DeleteWhenTimeOut(key, timer)
}

func (mc *MemCache) DeleteWhenTimeOut(key string, timer *time.Timer) {

	for {
		select {
		case <-timer.C:
			mc.timers.Delete(key)
			mc.cacheObject.Delete(key)
			return
		}
	}
}

func (mc *MemCache) ResetTimer(key string, expirationLimit int64) {
	timerInterface, ok := mc.timers.Load(key)
	if ok {
		timer := timerInterface.(*time.Timer)
		timer.Reset(time.Second * time.Duration(expirationLimit))
	}
}

func (mc *MemCache) Add(key string, value any, expirationLimit int64) {
	mc.cacheObject.Store(key, value)
	mc.CreateTimerToDelete(key, expirationLimit)
}

func (mc *MemCache) Del(key string) {
	timerInterface, ok := mc.timers.Load(key)
	if ok {
		timer := timerInterface.(*time.Timer)
		timer.Stop()
		mc.timers.Delete(key)
	}
	mc.cacheObject.Delete(key)
}

func (mc *MemCache) Get(key string) (any, bool) {
	return mc.cacheObject.Load(key)
}

func Add(key string, value any, expirationLimit int64) {
	memCache.Add(key, value, expirationLimit)
}

func Get(key string) (any, bool) {
	return memCache.Get(key)
}

func Del(key string) {
	memCache.Del(key)
}

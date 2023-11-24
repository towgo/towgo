/*
请求阻拦器
可用于限制用户的每秒最大请求数等场景
by:liangliangit
*/
package requestblocker

import (
	"sync"
	"time"
)

type RequestBlocker struct {
	sync.Mutex
	sessionFeaturesTimer map[string]*time.Timer
	sessionFeatures      map[string]*RequestBlock
	ResetTimeleft        int64 //CD冷却锁定时间
	BlockCountLimit      int64 //请求锁定阀值
}
type RequestBlock struct {
	RequestTimeLast int64 //最近一次的请求时间
	RequestCount    int64 //请求次数
	BlockCountLimit int64 //请求数量限定次数
	BlockTime       int64 //被锁定的时间点
	time.Timer
}

func New(ResetTimeleft, BlockCountLimit int64) *RequestBlocker {
	return &RequestBlocker{
		sessionFeaturesTimer: map[string]*time.Timer{},
		sessionFeatures:      map[string]*RequestBlock{},
		ResetTimeleft:        ResetTimeleft,
		BlockCountLimit:      BlockCountLimit,
	}
}

func (l *RequestBlocker) getFeature(feature string) *RequestBlock {
	l.Lock()
	defer l.Unlock()
	return l.sessionFeatures[feature]
}
func (l *RequestBlocker) setFeature(feature string) {
	l.Lock()

	l.sessionFeatures[feature] = &RequestBlock{
		BlockCountLimit: l.BlockCountLimit, //请求次数
	}

	//创建定时器用于删除请求特征
	timer := time.NewTimer(time.Second * time.Duration(l.ResetTimeleft))
	l.sessionFeaturesTimer[feature] = timer

	l.Unlock()

	go func() {
		for {
			select {
			case <-timer.C:
				l.removeFeature(feature)
			}
		}
	}()

}

func (l *RequestBlocker) resetTimer(feature string) {
	l.Lock()
	defer l.Unlock()
	l.sessionFeaturesTimer[feature].Reset(time.Second * time.Duration(l.ResetTimeleft))
}

func (l *RequestBlocker) removeFeature(feature string) {
	l.Lock()
	defer l.Unlock()
	delete(l.sessionFeatures, feature)
	delete(l.sessionFeaturesTimer, feature)

}

func (l *RequestBlocker) checkRequest(feature string) bool {

	session := l.getFeature(feature)
	if session == nil {
		return false
	}
	return session.check()
}

func (l *RequestBlocker) Request(feature string) bool {
	session := l.getFeature(feature)
	if session == nil {
		l.setFeature(feature)
	}
	l.resetTimer(feature)
	return l.checkRequest(feature)
}

func (rb *RequestBlock) check() bool {
	nowRequestTime := time.Now().Unix()
	//请求没有被限制
	if rb.RequestCount < rb.BlockCountLimit {
		rb.RequestCount++
		rb.RequestTimeLast = nowRequestTime
		return true
	}

	//请求超过阀值

	//检查是否被锁定
	if rb.BlockTime == 0 { //没有被锁定  进行锁定
		rb.RequestTimeLast = nowRequestTime
		rb.BlockTime = nowRequestTime
		return false
	}

	rb.RequestTimeLast = nowRequestTime
	return false
}

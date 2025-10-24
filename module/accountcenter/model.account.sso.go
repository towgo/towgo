package accountcenter

import (
	"errors"
	"log"
	"time"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/system"
)

var SsoTokenExpiredTime int64 = 10
var SsoAutoClearLimit int64 = 60 * 60 * 24 * 1

func (t *SSOToken) TableName() string {
	return "sso_token"
}

// SSOToken SSO token
type SSOToken struct {
	ID      int64  `json:"id"`
	Token   string `json:"token"`
	Uid     int64  `json:"uid"`
	Expired int64  `json:"expired"`
	User    any    `json:"user" xorm:"-" gorm:"-"`
}

func autoSsoTokenExpireTimeToClear() {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Print(err)
			}
			autoTimeToClear()
		}()
		var SsoToken SSOToken
		for {
			basedboperat.SqlExec("delete from "+SsoToken.TableName()+" where expired < ?", time.Now().Unix())
			time.Sleep(time.Second * time.Duration(SsoAutoClearLimit))
		}
	}()
}

// 验证是否过期
func (t *SSOToken) isValid() bool {
	return t.Expired > time.Now().Unix()
}

// NewSSOToken 根据 session 创建token
func NewSSOToken(session string) (*SSOToken, error) {
	token, err := GetToken(session)
	if err != nil {
		return nil, err
	}
	var ssoToken SSOToken
	if token.Uid == 0 {
		return nil, errors.New("code不存在或已过期") //数据不存在
	}
	guid := system.GetGUID().Hex()
	ssoToken.Token = system.MD5(guid)
	ssoToken.Uid = token.Uid
	ssoToken.Expired = time.Now().Unix() + SsoTokenExpiredTime //10秒
	ssoToken.User = token.Payload
	_, err = basedboperat.Create(&ssoToken)
	if err != nil {
		log.Println("error:", err)
		return nil, err
	}
	memCache.Add(ssoToken.Token, &ssoToken)
	return &ssoToken, nil
}

// GetSSOToken 根据tokenKey获取token
func GetSSOToken(tokenKey string) (*SSOToken, error) {
	var ssoToken = &SSOToken{}

	//缓存查询
	userTokenInterface, ok := memCache.Get(tokenKey)
	if ok {
		//缓存命中
		ssoToken = userTokenInterface.(*SSOToken)
		//缓存过期  清理
		if !ssoToken.isValid() {
			memCache.Del(tokenKey)
			return nil, errors.New("code 已过期")
		}
	} else {
		//数据库查询

		//查询持久化数据
		err := basedboperat.Get(ssoToken, nil, "token = ?", tokenKey)
		if err != nil {
			return nil, err
		}
		if ssoToken.Uid == 0 {
			return nil, errors.New("code不存在或已过期") //数据不存在
		}

		//token过期
		if !ssoToken.isValid() {
			return nil, errors.New("code不存在或已过期")
		}
		var user = &Account{}
		basedboperat.Get(user, nil, "id = ?", ssoToken.Uid)
		if user.ID == 0 {
			return nil, errors.New("code关联用户不存在") //用户不存在
		}
		ssoToken.User = user
		memCache.Add(ssoToken.Token, ssoToken)
	}

	return ssoToken, nil
}

func deleteSsoToken(code string) {
	memCache.Del(code)
	basedboperat.Delete(&SSOToken{}, nil, "token = ?", code)
}

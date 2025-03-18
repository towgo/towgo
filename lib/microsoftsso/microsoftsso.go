package microsoftsso

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

type SSOConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AuthUrl      string
	TokenUrl     string
}

type UserInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Mail        string `json:"mail"`
}

var oauthConfig = &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	RedirectURL:  redirectURI,
	Scopes:       []string{"openid", "email", "profile", "User.Read"},
	Endpoint:     microsoft.AzureADEndpoint("common"),
}

var clientID = "a6723367-d34c-4a91-9406-6b6bcecb604f"
var clientSecret = "62c4ddd0-75d7-4893-8b4f-baf16d398f37"
var redirectURI = "http://cntsns706.fle01.flender.net:19000/ssocallback"
var authUrl = "https://login.microsoftonline.com/3ab4e0b8-fd49-4b7d-819b-b92b8e5fb6a1/oauth2/v2.0/authorize"
var tokenUrl = "https://login.microsoftonline.com/3ab4e0b8-fd49-4b7d-819b-b92b8e5fb6a1/oauth2/v2.0/token"

func Init(config SSOConfig) {
	clientID = config.ClientID
	clientSecret = config.ClientSecret
	redirectURI = config.RedirectURI
	authUrl = config.AuthUrl
	tokenUrl = config.TokenUrl

}

// 获取访问令牌
func GetAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", tokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))))
	resp, err := httpClient.Do(req)
	// resp, err := http.PostForm(tokenUrl, data)
	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err.Error())
	}

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	log.Println("get token:", string(b))
	err = json.Unmarshal(b, &result)
	if err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

func GoAuth(w http.ResponseWriter, r *http.Request) {
	// https://idp.cloud.vwgroup.com/auth/realms/kums-mfa/protocol/openid-connect/auth?response_type=code&scope=openid+profile+email&client_id=idp-57f2c70a-c203-4d1f-a002-ab86bd95792b-Hub-RPACOEPlatform-Prod&redirect_uri=https%3A%2F%2Fcoeplatform.rpa.prod.vwah.vwgroup.com/*%2Fcallback
	authURL := fmt.Sprintf(authUrl+"?client_id=%s&response_type=code&redirect_uri=%s&scope=openid+profile+email", clientID, url.QueryEscape(redirectURI))
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// 获取用户信息
func GetUserName(token string) (*UserInfo, error) {
	secret := []byte("key")

	// tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJJQjNRTk5TWUJibUJ6ajNaRTM4aFY3dWZYUWhWbjlUcWhDUFpfWlhDTDFnIn0.eyJleHAiOjE3MjczNDM3MjgsImlhdCI6MTcyNzM0MTkyOCwiYXV0aF90aW1lIjoxNzI3MzQwNTEzLCJqdGkiOiI4OWRjZTA1ZS01Mzk2LTRmZjktYTQ2Yi0yMjQ1OTRiOGY4NDUiLCJpc3MiOiJodHRwczovL2lkcC5jbG91ZC52d2dyb3VwLmNvbS9hdXRoL3JlYWxtcy9rdW1zLW1mYSIsImF1ZCI6ImlkcC01N2YyYzcwYS1jMjAzLTRkMWYtYTAwMi1hYjg2YmQ5NTc5MmItSHViLVJQQUNPRVBsYXRmb3JtLVByb2QiLCJzdWIiOiJiMjIxZWVlNC0yMDNhLTRjYjctOGFkNi0yNjUyYTY3MDE4YzAiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJpZHAtNTdmMmM3MGEtYzIwMy00ZDFmLWEwMDItYWI4NmJkOTU3OTJiLUh1Yi1SUEFDT0VQbGF0Zm9ybS1Qcm9kIiwic2Vzc2lvbl9zdGF0ZSI6ImJhODMwYTZmLTQ0NzAtNGUyMi05YmU0LTExY2RjNDQwYzA1YSIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJzaWQiOiJiYTgzMGE2Zi00NDcwLTRlMjItOWJlNC0xMWNkYzQ0MGMwNWEiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkNoYW5ncWluZyBMaXUiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJ2ZnB2a2Z0IiwiZ2l2ZW5fbmFtZSI6IkNoYW5ncWluZyIsImZhbWlseV9uYW1lIjoiTGl1IiwiZW1haWwiOiJleHRlcm4uY2hhbmdxaW5nLmxpdUB2b2xrc3dhZ2VuLWFuaHVpLmNvbSJ9.WivBfJx9ljDnbTXKox7ZmqjsvVEcGQd1c-74BabX2-g6Zz9Mtfx9HRKK27YEY3IRqb-Czjp4uHkqBGLP1qdT2wWaSR1dynKEVnWozuJ78yKiM3BsDsyotXq6wWgsYIKB-mbK6-HjG4grlL18C5YGHfXdnymzZJ2pznDo9Bc8_JgSYG9RtO8Wq-Uz8TEtPaVtnIeT2l_eRIhoo3AFN1nqCZRwjyfsTVm7S0OMZhgkoBPOXnvbsxKokYMQ7bxrloUV3UAsMPXHMYrTnzkb1ULRvarBrSx3sDpuGvXAMIMWPJEUhatU70LEx5zy5rDgYj8xf7nfwo8qGxrLuc1uTrTI5A"
	// log.Println("token:",token)
	jwtToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("不是jwt.MapClaims对象")
	}
	// req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
	// if err != nil {
	// 	return nil, err
	// }

	// req.Header.Set("Authorization", "Bearer "+token)

	// resp, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	return nil, err
	// }
	// defer resp.Body.Close()

	// b, err := io.ReadAll(resp.Body)

	// if err != nil {
	// 	log.Print(err.Error())
	// }

	var result UserInfo
	// 客户要求统一转为大写
	result.Username = strings.ToUpper(claims["preferred_username"].(string))
	result.DisplayName = claims["name"].(string)
	result.Mail = claims["email"].(string)
	return &result, nil
}

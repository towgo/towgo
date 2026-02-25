package accountcenter

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/towgo/towgo/lib/jsonrpc"
)

// 初始化SSO API
func initSSOApi() {
	http.HandleFunc("/account/sso/code", ssoCode)
	http.HandleFunc("/account/sso/auth", ssoAuth)
	go autoSsoTokenExpireTimeToClear()
}

func ssoCode(writer http.ResponseWriter, request *http.Request) {
	session := request.URL.Query().Get("session")
	if session == "" {
		http.Error(writer, "session is empty", http.StatusBadRequest)
		return
	}
	callback := request.URL.Query().Get("callback")
	if callback == "" {
		http.Error(writer, "callback is empty", http.StatusBadRequest)
		return
	}

	token, err := NewSSOToken(session)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(writer, request, callback+"?code="+token.Token, http.StatusTemporaryRedirect)
}

func ssoAuth(writer http.ResponseWriter, request *http.Request) {
	var restful jsonrpc.HttpRestfulResult

	bytes, err := io.ReadAll(request.Body)

	if err != nil {
		restful.Status = 500
		restful.Message = err.Error()
		result, _ := json.Marshal(restful)
		writer.Write(result)
		return
	}

	var params struct {
		Code string `json:"code"`
	}
	err = json.Unmarshal(bytes, &params)
	if err != nil {
		restful.Status = 500
		restful.Message = err.Error()
		result, _ := json.Marshal(restful)
		writer.Write(result)
		return
	}
	if params.Code == "" {
		restful.Status = 500
		restful.Message = "code 必填"
		result, _ := json.Marshal(restful)
		writer.Write(result)
		return
	}
	token, err := GetSSOToken(params.Code)
	if err != nil {
		restful.Status = 500
		restful.Message = err.Error()
		result, _ := json.Marshal(restful)
		writer.Write(result)
		return
	}
	if !token.isValid() {
		restful.Status = 500
		restful.Message = "code 已过期"
		result, _ := json.Marshal(restful)
		writer.Write(result)
		return
	}
	deleteSsoToken(params.Code)
	restful.Status = 200
	account := token.User.(*Account)
	restful.Result = map[string]interface{}{
		"id":         account.ID,
		"username":   account.Username,
		"nickname":   account.Nickname,
		"created_at": account.CreatedAt,
		"updated_at": account.UpdatedAt}
	result, _ := json.Marshal(restful)
	writer.Write(result)
}

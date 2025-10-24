package accountcenter

import (
	"github.com/goccy/go-json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/towgo/towgo/lib/microsoftsso"
	"github.com/towgo/towgo/lib/system"
)

var ssologin_template_path string

func init() {

	http.HandleFunc("/ssologin", microsoftsso.GoAuth)

	// 启动一个本地服务器，用于接收回调
	http.HandleFunc("/ssocallback", ssocallback)
	// 启动一个本地服务器，用于接收回调
	http.HandleFunc("/ssoMicoLogin", ssoMicoLogin)
}

func ssocallback(w http.ResponseWriter, r *http.Request) {
	// 获取授权码
	code := r.URL.Query().Get("code")

	// 使用授权码获取访问令牌
	token, err := microsoftsso.GetAccessToken(code)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	if token == "" {
		microsoftsso.GoAuth(w, r)
		return
	}
	log.Println("ssotoken:---", token)
	// 使用访问令牌获取用户信息
	userinfo, err := microsoftsso.GetUserName(token)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	var account Account
	token = system.SHA1(token)
	err = account.RegAndUpdateOauth(userinfo.Username, userinfo.DisplayName, userinfo.Mail, token, userinfo.ID)

	// if err != nil {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	var data struct {
		Username string
		Password string
		Msg      string
	}
	if err != nil {
		data.Msg = err.Error()
		data.Password = ""
	} else {
		data.Password = token
	}
	data.Username = userinfo.Username
	t1, err := template.ParseFiles("./template/sso_login.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	t1.Execute(w, data)

}

func ssoMicoLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只接受POST请求", http.StatusMethodNotAllowed)
		return
	}
	// 读取POST请求体数据
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusInternalServerError)
		return
	}
	// 获取授权码
	var param struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	/*param.Username = r.URL.Query().Get("username")
	param.Password = r.URL.Query().Get("password")*/
	err = json.Unmarshal(body, &param)
	if err != nil {
		http.Error(w, "解析JSON数据失败", http.StatusBadRequest)
		return
	}
	var account Account
	param.Username = strings.ToLower(param.Username)
	loginErr := account.MicrosoftAdLogin(param.Username, param.Password)
	// if loginErr != nil { //模型层登录成功
	// 	// dblog.Write("account:login","", fmt.Sprintf("%s@%s 登录失败！ 错误信息:%s", account.Username, rpcConn.GetRemoteAddr(), loginErr.Error()))
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(loginErr.Error()))
	// 	return
	// }

	// // 使用授权码获取访问令牌
	// token, err := microsoftsso.GetAccessToken(code)
	// if err != nil {
	// 	log.Print(err)
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	// if token == ""{
	// 	microsoftsso.GoAuth(w,r)
	// 	return
	// }
	// log.Println("ssotoken:---",token)
	// // 使用访问令牌获取用户信息
	// userinfo, err := microsoftsso.GetUserName(token)
	// if err != nil {
	// 	log.Print(err)
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }

	// token := system.MD5(account.UserToken.TokenKey)
	// err = account.RegAndUpdateOauth(userinfo.Username, userinfo.DisplayName, userinfo.Mail, token, userinfo.ID)

	// if err != nil {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }
	var data struct {
		Username string
		Password string
		Msg      string
	}
	if loginErr != nil {
		data.Msg = loginErr.Error()
		data.Password = ""
	} else {
		data.Password = param.Password
	}
	data.Username = param.Username
	t1, err := template.ParseFiles("./template/sso_login.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	t1.Execute(w, data)

}

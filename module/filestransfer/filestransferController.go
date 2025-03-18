package filestransfer

import (
	"encoding/json"
	"github.com/towgo/towgo/module/dblog"
	"io"
	"log"
	"net/http"
	"path"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"github.com/towgo/towgo/lib/api"
	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/module/accountcenter"
	"github.com/towgo/towgo/module/accountcenter/accountctx"
)

var (
	http_download_path       string = "/filestransfer/download"
	http_upload_path         string = "/filestransfer/upload"
	http_upload_path_restful string = "/restful/filestransfer/upload"
	token_key                string = "token"
	token_cookies_key        string = "ctoken"
	auth                     bool   = true
	allowCross               bool   = false
)

func Auth(b bool) {
	auth = b
}

func AllowCross(b bool) {
	allowCross = b
}

func SetApiHeader(header string) {
	http_download_path = header + http_download_path
	http_upload_path = header + http_upload_path
	http_upload_path_restful = header + http_upload_path_restful
}

func SetTokenKey(key string) {
	token_key = key
}

func InitApi() {
	http.HandleFunc(http_download_path, http_filestransfer_download)
	http.HandleFunc(http_upload_path, http_filestransfer_upload)
	http.HandleFunc(http_upload_path_restful, http_filestransfer_upload_restful)
	jsonrpc.SetFunc("/filestransfer/upload", jsonrpc_filestransfer_create)

	xormDriver.Sync2(new(FileKey))
	api.NewCRUDJsonrpcAPI("/filestransfer", FileKey{}, []FileKey{}).RegAPI(api.CRUD_FLAG_LIST, api.CRUD_FLAG_DELETE, api.CRUD_FLAG_DETAIL)
	jsonrpc.SetFunc("/filestransfer/setOwner", jsonrpc_filestransfer_set_owner)
	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.UPLOAD, "/filestransfer/upload", "account_center", "文件上传"),
		dblog.NewOperateType(dblog.UPDATE, "/filestransfer/setOwner", "account_center", "设置文件所有者"),
		dblog.NewOperateType(dblog.QUERY, "/filestransfer/list", "account_center", "查看文件列表"),
		dblog.NewOperateType(dblog.DELETE, "/filestransfer/delete", "account_center", "删除文件"),
	)
}

func jsonrpc_filestransfer_set_owner(rpcConn jsonrpc.JsonRpcConnection) {

	account, err := accountctx.Parse(rpcConn)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
	}

	var params struct {
		Owners  []Account `json:"owners"`
		FileKey string    `json:"file_key"`
	}

	rpcConn.ReadParams(&params)

	var fileKey FileKey
	basedboperat.Get(&fileKey, nil, "file_key = ?", params.FileKey)

	if fileKey.FileKey == "" {
		rpcConn.WriteError(500, "文件不存在")
		return
	}

	if fileKey.Creator != account.Username {
		rpcConn.WriteError(500, "非创建用户不能修改权限")
		return
	}

	var owners []Account
	owners = append(owners, Account{ID: account.ID, Username: account.Username})
	for _, v := range params.Owners {
		if v.ID == account.ID {
			continue
		}
		owners = append(owners, v)
	}

	fileKey.Owners = owners

	err = basedboperat.Update(&fileKey, "owners", "file_key = ?", params.FileKey)
	if err != nil {
		rpcConn.WriteError(500, err.Error())
		return
	}

	rpcConn.WriteResult("ok")
}

func jsonrpc_filestransfer_create(rpcConn jsonrpc.JsonRpcConnection) {

	_, err := accountctx.Parse(rpcConn)
	if err != nil {
		rpcConn.GetRpcResponse().Error.Set(401, err.Error())
	}

	var files FilesObject
	rpcConn.ReadParams(&files)

	files.SaveAll()
	for k, _ := range files.Files {
		files.Files[k].Data = ""
	}
	rpcConn.WriteResult(files)
}

func http_filestransfer_download(w http.ResponseWriter, r *http.Request) {

	if allowCross {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}

	filekey := r.URL.Query().Get(queryStringKey)
	download := r.URL.Query().Get("download")
	filename := r.URL.Query().Get("filename")
	cookies := r.Cookies()
	ctoken := ""
	for _, v := range cookies {
		if v.Name == token_cookies_key {
			ctoken = v.Value
			break
		}
	}
	token := r.Header.Get(token_key)
	token2 := r.Header.Get("Token")
	endToken := ""
	if ctoken != "" {
		endToken = ctoken
	}
	if token != "" && endToken == "" {
		endToken = token
	}
	if token2 != "" && endToken == "" {
		endToken = token2
	}
	log.Println("ctoken = ", ctoken)
	log.Println("token = ", token)
	log.Println("token2 = ", token2)
	var account accountcenter.Account
	var resp RestFulResult
	if endToken == "" {
		resp.Message = "Image download fail , Unauthorized ,Please login  endToken = " + endToken
		resp.Status = 401
		response, _ := json.Marshal(resp)

		w.WriteHeader(401)
		w.Write(response)
		return
	}
	log.Println("jsonrpc endToken = ", endToken)
	jsonrpc.Call("/account/myinfo", endToken, nil, &account)
	if account.ID == 0 {
		resp.Message = "Image download fail，Account Unauthorized endToken = " + endToken
		resp.Status = 401
		response, _ := json.Marshal(resp)
		w.WriteHeader(401)
		w.Write(response)
		return
	}

	if filekey == "" {
		w.WriteHeader(403)
		w.Write([]byte("queryString filekey can not be null"))
		return
	}
	file := File{}
	file.FileKey = filekey
	_, err := file.GetDataStream()
	if err != nil {
		w.WriteHeader(403)
		log.Print(err.Error())
		w.Write([]byte("文件不存在或加载失败"))
		return
	}
	var contentType string

	if download == "true" {
		contentType = "application/octet-stream"
	} else {
		contentType = contentMap[file.Suffix]
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("content-type", contentType)

	//判断是否使用自定义文件名
	if filename == "" {
		filename = file.Name + file.Suffix
	}
	w.Header().Set("Content-Disposition", "filename="+filename)

	//向客户端写入文件（流写入）
	io.Copy(w, &file)

}

func http_filestransfer_upload(w http.ResponseWriter, r *http.Request) {
	if allowCross {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}
	if r.Method == "POST" {
		token := r.Header.Get(token_key)
		if token == "" {
			w.WriteHeader(401)
			w.Write([]byte("未授权,请登录"))
			return
		}

		var account accountcenter.Account
		jsonrpc.Call("/account/myinfo", token, nil, &account)
		if account.ID == 0 {
			w.WriteHeader(401)
			w.Write([]byte("账户未授权"))
			return
		}

		responseJson := struct {
			Success int    `json:"success"`
			Message string `json:"message"`
			Url     string `json:"url"`
			FileKey string `json:"filekey"`
		}{}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			responseJson.Success = 0
			responseJson.Message = err.Error()
			response, _ := json.Marshal(responseJson)
			w.Write(response)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			responseJson.Success = 0
			responseJson.Message = "未找到文件"
			response, _ := json.Marshal(responseJson)
			w.Write(response)
			return
		}
		defer file.Close()

		extName := path.Ext(handler.Filename)

		if len(allowUploadExt) > 0 {
			_, ok := allowUploadExt[extName]
			if !ok {
				responseJson.Success = 0
				responseJson.Message = extName + "文件不被允许上传"
				response, _ := json.Marshal(responseJson)
				w.Write(response)
				return
			}
		}

		f := &File{}

		f.OwnerUsers = append(f.OwnerUsers, account.ID)
		f.Suffix = extName
		f.Name = handler.Filename
		f.EncodeType = "binary"
		f.InitForSave(r.URL.Query().Get(queryStringKey))
		_, err = io.Copy(f, file)
		if err != nil {
			responseJson.Success = 0
			responseJson.Message = err.Error()
			response, _ := json.Marshal(responseJson)
			w.Write(response)
			return
		}

		//写入数据库
		var fileKey FileKey
		basedboperat.Get(&fileKey, nil, "file_key = ?", f.FileKey)

		fileKey.Name = f.Name
		fileKey.Creator = account.Username
		fileKey.DwonloadUrl = f.DwonloadUrl
		fileKey.UploadType = r.URL.Query().Get("upload_type")

		if fileKey.FileKey == "" {
			fileKey.FileKey = f.FileKey
			_, err := basedboperat.Create(&fileKey)
			if err != nil {
				responseJson.Success = 0
				responseJson.Message = err.Error()
				response, _ := json.Marshal(responseJson)
				w.Write(response)
				return
			}

		} else {
			err := basedboperat.Update(&fileKey, []string{"name", "creator", "download_url"}, "file_key = ?", f.FileKey)
			if err != nil {
				responseJson.Success = 0
				responseJson.Message = err.Error()
				response, _ := json.Marshal(responseJson)
				w.Write(response)
				return
			}
		}

		responseJson.Success = 1
		responseJson.Url = f.DwonloadUrl
		responseJson.FileKey = f.FileKey
		response, _ := json.Marshal(responseJson)
		w.Write(response)
		return
	}

	w.WriteHeader(403)
	w.Write([]byte("bad request for upload"))
}

func http_filestransfer_upload_restful(w http.ResponseWriter, r *http.Request) {
	if allowCross {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
	}
	var resp RestFulResult

	if r.Method == "POST" {
		var account accountcenter.Account

		//需要身份认证
		if auth {
			token := r.Header.Get(token_key)
			if token == "" {
				resp.Message = "未授权,请登录"
				resp.Status = 401
				response, _ := json.Marshal(resp)

				w.WriteHeader(401)
				w.Write(response)
				return
			}

			jsonrpc.Call("/account/myinfo", token, nil, &account)

			if account.ID == 0 {
				resp.Message = "账户未授权"
				resp.Status = 401
				response, _ := json.Marshal(resp)
				w.WriteHeader(401)
				w.Write(response)
				return
			}
		} else {
			account.ID = -1
			account.Username = "guest"
		}

		responseJson := struct {
			Success int    `json:"success"`
			Message string `json:"message"`
			Url     string `json:"url"`
			FileKey string `json:"filekey"`
		}{}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			resp.Message = err.Error()
			resp.Status = 500
			response, _ := json.Marshal(resp)
			w.WriteHeader(500)
			w.Write(response)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			resp.Message = "file字段不存在"
			resp.Status = 500
			response, _ := json.Marshal(resp)
			w.WriteHeader(500)
			w.Write(response)
			return
		}
		defer file.Close()
		if handler.Size == 0 {
			resp.Message = "文件为空无法上传"
			resp.Status = 500
			response, _ := json.Marshal(resp)
			w.WriteHeader(500)
			w.Write(response)
			return
		}
		extName := path.Ext(handler.Filename)

		if len(allowUploadExt) > 0 {
			_, ok := allowUploadExt[extName]
			if !ok {
				responseJson.Success = 0
				responseJson.Message = extName + "文件不被允许上传"
				response, _ := json.Marshal(responseJson)
				w.Write(response)
				return
			}
		}

		f := &File{}
		f.OwnerUsers = append(f.OwnerUsers, account.ID)
		f.Suffix = extName
		f.Name = handler.Filename
		f.EncodeType = "binary"
		f.InitForSave(r.URL.Query().Get(queryStringKey))

		_, err = io.Copy(f, file)
		if err != nil {
			resp.Message = err.Error()
			resp.Status = 500
			response, _ := json.Marshal(resp)
			w.WriteHeader(500)
			w.Write(response)
			return
		}

		//写入数据库
		var fileKey FileKey
		basedboperat.Get(&fileKey, nil, "file_key = ?", f.FileKey)

		fileKey.Name = f.Name
		fileKey.Creator = account.Username
		fileKey.DwonloadUrl = f.DwonloadUrl
		fileKey.UploadType = r.URL.Query().Get("upload_type")
		if fileKey.FileKey == "" {
			fileKey.FileKey = f.FileKey
			_, err := basedboperat.Create(&fileKey)
			if err != nil {
				resp.Message = err.Error()
				resp.Status = 500
				response, _ := json.Marshal(resp)
				w.WriteHeader(500)
				w.Write(response)
				return
			}

		} else {
			err := basedboperat.Update(&fileKey, []string{"name", "creator", "download_url"}, "file_key = ?", f.FileKey)
			if err != nil {
				resp.Message = err.Error()
				resp.Status = 500
				response, _ := json.Marshal(resp)
				w.WriteHeader(500)
				w.Write(response)
				return
			}
		}

		responseJson.Success = 1
		responseJson.Url = f.DwonloadUrl
		responseJson.FileKey = f.FileKey

		resp.Message = "ok"
		resp.Status = 200
		resp.Result = responseJson

		response, _ := json.Marshal(resp)
		w.Write(response)
		return
	}
	resp.Message = "bad request for upload"
	resp.Status = 403
	response, _ := json.Marshal(resp)
	w.WriteHeader(403)
	w.Write(response)
}

package www

import (
	"net/http"
	"os"
	"path"
)

type WebServer struct {
	Wwwroot string
	Index   []string
}

func (webserver *WebServer) WebServerHandller(w http.ResponseWriter, r *http.Request) {

	var wwwroot string = webserver.Wwwroot
	filepath := wwwroot + r.URL.Path

	var f []byte
	var err error

	for _, v := range webserver.Index {
		if r.URL.Path == "/" {
			filepath = wwwroot + "/" + v
		}
		f, err = os.ReadFile(filepath)
		if err == nil {
			break
		}
	}

	if err != nil {
		for _, v := range webserver.Index {
			filepath = wwwroot + "/" + v
			f, err = os.ReadFile(filepath)
			if err == nil {
				webserver.WriteFile(w, filepath, f)
				return

			}
		}

		w.WriteHeader(404)
		w.Write([]byte("404 - " + http.StatusText(404)))
		return
	}

	webserver.WriteFile(w, filepath, f)

}

func (webserver *WebServer) WriteFile(w http.ResponseWriter, filepath string, f []byte) {
	var contentType string

	extName := path.Ext(filepath)
	switch extName {
	case ".css":
		contentType = "text/css"
	case ".html":
		contentType = "text/html"
	case ".js":
		contentType = "application/javascript"
	case ".png":
		contentType = "image/png"
	case ".jpg":
		contentType = "image/jpeg"
	case ".jpeg":
		contentType = "image/jpeg"
	case ".svg":
		contentType = "image/svg+xml"
	case ".json":
		contentType = "application/json"
	case ".gif":
		contentType = "image/gif"
	case ".zip":
		contentType = "application/x-zip-compressed"
	case ".rar":
		contentType = "application/octet-stream"
	case ".txt":
		contentType = "text/plain"
	case ".pdf":
		contentType = "application/pdf"
	case ".doc":
		contentType = "application/msword"
	case ".docx":
		contentType = "application/msword"
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(f)
}

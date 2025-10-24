package www

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type WebServer struct {
	Wwwroot    string   `json:"wwwroot"`
	Index      []string `json:"index"`
	EmbeddedFS *embed.FS
}

func (webserver *WebServer) WebServerHandller(w http.ResponseWriter, r *http.Request) {

	var wwwroot string = webserver.Wwwroot
	filepath := wwwroot + r.URL.Path

	filepath = ConvertPathSeparator(filepath)

	isFile, err := isFilePath(r.URL.Path)
	if err != nil {
		log.Print(err.Error())
	}

	var f []byte

	if isFile {
		if webserver.EmbeddedFS != nil {
			f, err = webserver.EmbeddedFS.ReadFile(filepath)
			if err != nil {
				f, err = os.ReadFile(filepath)
			}
		} else {
			f, err = os.ReadFile(filepath)
		}
	} else {
		//路径模式直接判断命中
		path := ensureTrailingSlash(r.URL.Path)
		for _, v := range webserver.Index {
			filepath = wwwroot + path + v

			if webserver.EmbeddedFS != nil {
				f, err = webserver.EmbeddedFS.ReadFile(filepath)
				if err != nil {
					f, err = os.ReadFile(filepath)
				}
			} else {
				f, err = os.ReadFile(filepath)
			}

			if err == nil {
				break
			}
		}

		//路径模式下 没有直接命中,进行目录递归查找默认首页
		if err != nil {
			iter := NewPathIterator(path)
			for p := iter.Next(); p != ""; p = iter.Next() {
				for _, v := range webserver.Index {
					filepath = wwwroot + ensureTrailingSlash(p) + v

					if webserver.EmbeddedFS != nil {
						f, err = webserver.EmbeddedFS.ReadFile(filepath)
						if err != nil {
							f, err = os.ReadFile(filepath)
						}
					} else {
						f, err = os.ReadFile(filepath)
					}

					if err == nil {
						break
					}
				}
				if err == nil {
					break
				}
			}
		}

	}

	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte("404 - " + http.StatusText(404)))
		return
	}

	webserver.WriteFile(w, r, filepath, f)

}

func (webserver *WebServer) WriteFile(w http.ResponseWriter, r *http.Request, filepath string, f []byte) {
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
	case ".mp4":
		videoHandler(filepath, w, r)
		return
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(f)
}

func videoHandler(filepath string, w http.ResponseWriter, r *http.Request) {
	// Specify the path to your video file

	file, err := os.Open(filepath)
	if err != nil {
		http.Error(w, "Could not open video file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get video file stat", http.StatusInternalServerError)
		return
	}

	fileSize := fileStat.Size()
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		// If there is no Range header, serve the whole file
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		http.ServeFile(w, r, filepath)
		return
	}

	// Extract the range start and end
	rangeParts := strings.Split(rangeHeader, "=")
	if len(rangeParts) != 2 {
		http.Error(w, "Invalid Range header", http.StatusBadRequest)
		return
	}

	rangeValues := strings.Split(rangeParts[1], "-")
	if len(rangeValues) != 2 {
		http.Error(w, "Invalid Range header", http.StatusBadRequest)
		return
	}

	start, err := strconv.ParseInt(rangeValues[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid Range start value", http.StatusBadRequest)
		return
	}

	var end int64
	if rangeValues[1] == "" {
		end = fileSize - 1
	} else {
		end, err = strconv.ParseInt(rangeValues[1], 10, 64)
		if err != nil {
			http.Error(w, "Invalid Range end value", http.StatusBadRequest)
			return
		}
	}

	if start > end || end >= fileSize {
		http.Error(w, "Invalid Range values", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// Set the headers for partial content
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Seek to the start position and stream the desired range
	file.Seek(start, 0)
	buffer := make([]byte, 1024)
	for {
		readBytes := end - start + 1
		if readBytes < int64(len(buffer)) {
			buffer = buffer[:readBytes]
		}
		n, err := file.Read(buffer)
		if err != nil || n == 0 {
			break
		}
		w.Write(buffer[:n])
		start += int64(n)
	}
}

func ensureTrailingSlash(s string) string {
	// 特殊情况处理
	if s == "" {
		return "/"
	}

	// 从字符串末尾向前遍历，跳过所有连续的斜杠
	i := len(s) - 1
	for i >= 0 && s[i] == '/' {
		i--
	}

	// 处理不同情况
	switch {
	case i < 0: // 全部字符都是斜杠
		return "/"
	case i == len(s)-1: // 最后字符不是斜杠
		return s + "/"
	default: // 有斜杠但已经跳过末尾多余斜杠
		return s[:i+1] + "/"
	}
}

// isFilePath 判断 URL 路径是否是文件
// 规则：如果路径最后一部分包含点号（.），则认为是文件，否则是路径
func isFilePath(urlStr string) (bool, error) {
	// 解析 URL 获取路径部分
	u, err := url.Parse(urlStr)
	if err != nil {
		return false, err
	}

	// 获取清理后的路径（去除多余斜杠等）
	cleanPath := path.Clean(u.Path)

	// 处理特殊路径情况
	switch cleanPath {
	case "", ".", "/":
		return false, nil // 根路径或当前路径视为目录
	}

	// 提取最后一部分路径元素
	lastSegment := path.Base(cleanPath)

	// 最后一部分包含点号则视为文件
	if containsDot(lastSegment) {
		return true, nil
	}

	// 没有点号则是目录路径
	return false, nil
}

// containsDot 判断字符串是否包含点号（.）
func containsDot(s string) bool {
	// 排除特殊目录名称
	if s == "." || s == ".." {
		return false
	}
	return strings.Contains(s, ".")
}

func ConvertPathSeparator(path string) string {
	// 判断当前操作系统
	if runtime.GOOS == "windows" {
		// Windows系统：将所有Linux路径分隔符(/)替换为Windows分隔符(\)
		return strings.ReplaceAll(path, "/", "\\")
	} else {
		// Linux/Unix系统：将所有Windows路径分隔符(\)替换为Linux分隔符(/)
		return strings.ReplaceAll(path, "\\", "/")
	}
}

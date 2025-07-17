package system

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var pathSymbol string

func init() {
	switch runtime.GOOS {
	case "windows":
		pathSymbol = "\\"
	default:
		pathSymbol = "/"
	}
}
func GetPathSymbol() string {
	return pathSymbol
}

func ReplaceRelativePath(s string) string {
	if len(s) >= 2 && (strings.HasPrefix(s, "./") || strings.HasPrefix(s, ".\\")) {
		return GetPathOfProgram() + s[2:]
	}
	return s
}

func GetPathOfProgram() string {
	if !IsRelease() {
		return "." + GetPathSymbol()
	}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径 filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", GetPathSymbol(), -1) + GetPathSymbol() //将\替换成/
}

func IsRelease() bool {
	arg1 := strings.ToLower(os.Args[0])
	return !strings.Contains(arg1, "go-build")
}

package system

import (
	"net"
	"net/http"
)

func RemoteIp(r *http.Request) string {

	if r.Header.Get("X-Forwarded-For") != "" {
		return r.Header.Get("X-Forwarded-For")
	}

	if r.Header.Get("X-Real-IP") != "" {
		return r.Header.Get("X-Real-IP")
	}

	if r.Header.Get("Ali-CDN-Real-IP") != "" {
		return r.Header.Get("Ali-CDN-Real-IP")
	}

	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}

	return "unkonw"
}

func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

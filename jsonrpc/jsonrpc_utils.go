package jsonrpc

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// cleanPrefix normalizes a router group prefix.
func cleanPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return ""
	}
	return cleanMethod(prefix)
}

// cleanMethod normalizes a JSON-RPC method path.
func cleanMethod(method string) string {
	method = strings.TrimSpace(method)
	if method == "" || method == "/" {
		return "/"
	}
	if !strings.HasPrefix(method, "/") {
		method = "/" + method
	}
	for strings.Contains(method, "//") {
		method = strings.ReplaceAll(method, "//", "/")
	}
	if len(method) > 1 {
		method = strings.TrimRight(method, "/")
	}
	return method
}

// joinMethod joins a group prefix and method path.
func joinMethod(prefix, method string) string {
	prefix = cleanPrefix(prefix)
	method = cleanMethod(method)
	if prefix == "" {
		return method
	}
	if method == "/" {
		return prefix
	}
	return cleanMethod(prefix + "/" + strings.TrimLeft(method, "/"))
}

// newConnectionGUID creates a random connection identifier.
func newConnectionGUID(prefix string) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return prefix + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return prefix + hex.EncodeToString(b)
}

// randomString returns a cryptographically random alpha-numeric string.
func randomString(n int) string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	for i := range buf {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			buf[i] = alphabet[time.Now().UnixNano()%int64(len(alphabet))]
			continue
		}
		buf[i] = alphabet[idx.Int64()]
	}
	return string(buf)
}

// timestampMillis returns the current Unix timestamp in milliseconds.
func timestampMillis() string {
	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}

// remoteIP resolves the best client IP from an HTTP request.
func remoteIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if ip := firstHeaderIP(r.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// firstHeaderIP returns the first IP in a comma-separated header value.
func firstHeaderIP(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

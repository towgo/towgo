package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

// ServeJSON executes a JSON-RPC request encoded as bytes and returns response bytes.
func (s *JsonrpcServer) ServeJSON(ctx context.Context, data []byte) ([]byte, error) {
	var rpcReq *Jsonrpcrequest
	if len(data) == 0 {
		rpcReq = NewJsonrpcrequest()
	} else {
		var err error
		rpcReq, err = ToJsonrpcrequest(string(data))
		if err != nil {
			return nil, err
		}
	}
	rpcResp := s.Exec(ctx, rpcReq)
	return json.Marshal(rpcResp)
}

// ServeHTTP serves JSON-RPC over a standard net/http endpoint.
func (s *JsonrpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeHTTPError(w, 500, err.Error())
		return
	}
	if _, err := ToJsonrpcrequest(string(body)); err != nil {
		writeHTTPError(w, -32700, err.Error())
		return
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	conn := NewHttpRpcConnection(w, r)
	if conn == nil {
		return
	}
	conn.WithContext(r.Context())
	s.ServeConnection(conn)
}

// GhttpHandler serves JSON-RPC over a GoFrame ghttp route.
func (s *JsonrpcServer) GhttpHandler(r *ghttp.Request) {
	s.ServeHTTP(r.Response.Writer, r.Request)
}

// writeHTTPError writes a JSON-RPC error payload to an HTTP response.
func writeHTTPError(w http.ResponseWriter, code int64, msg string) {
	rpcResp := NewJsonrpcresponse()
	rpcResp.Error.Set(code, msg)
	_ = json.NewEncoder(w).Encode(rpcResp)
}

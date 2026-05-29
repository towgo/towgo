package jsonrpc_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/v2/jsonrpc"
	"net/http"
	"net/http/httptest"
	"strings"
)

// ExampleJsonrpcServer_goframeStyle demonstrates GoFrame-style binding over common transports.
func ExampleJsonrpcServer_goframeStyle() {
	s := newDemoServer()

	rpcReq := jsonrpc.NewJsonrpcrequest()
	rpcReq.Method = "/api/demo/hello"
	rpcReq.Params = map[string]any{"name": "exec"}
	rpcResp := s.Exec(context.Background(), rpcReq)
	var execResult demoHelloRes
	mustDecodeExampleResult(rpcResp.Result, &execResult)
	fmt.Println("exec:", execResult.Message)

	jsonBytes, _ := s.ServeJSON(
		context.Background(),
		[]byte(`{"jsonrpc":"2.0","method":"/api/demo/hello","params":{"name":"json"},"id":"1"}`),
	)
	var jsonResp struct {
		Result demoHelloRes `json:"result"`
	}
	_ = json.Unmarshal(jsonBytes, &jsonResp)
	fmt.Println("json:", jsonResp.Result.Message)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(
		http.MethodPost,
		"/jsonrpc",
		strings.NewReader(`{"jsonrpc":"2.0","method":"/api/demo/hello","params":{"name":"http"},"id":"2"}`),
	)
	s.ServeHTTP(rr, req)
	var httpResp struct {
		Result demoHelloRes `json:"result"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &httpResp)
	fmt.Println("http:", httpResp.Result.Message)

	localReq := jsonrpc.NewJsonrpcrequest()
	localReq.Method = "/api/demo/hello"
	localReq.Params = map[string]any{"name": "local"}
	localResp := s.ExecConnection(jsonrpc.NewLocalRpcConnection(localReq, nil))
	var localResult demoHelloRes
	mustDecodeExampleResult(localResp.Result, &localResult)
	fmt.Println("connection:", localResult.Message)

	// Output:
	// exec: hello exec
	// json: hello json
	// http: hello http
	// connection: hello local
}

// mustDecodeExampleResult decodes an example response result without noisy error handling.
func mustDecodeExampleResult(value any, dest any) {
	data, _ := json.Marshal(value)
	_ = json.Unmarshal(data, dest)
}

package jsonrpc_test

import (
	"context"
	"encoding/json"
	"github.com/towgo/towgo/v2/jsonrpc"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

// traceKey stores middleware trace data in context.
type traceKey struct{}

// ctxIsolationKey verifies runtime context isolation from protocol requests.
type ctxIsolationKey struct{}

// demoController provides a GoFrame-style bound controller for tests.
type demoController struct{}

// demoHelloReq is the bound request type for demoController.Hello.
type demoHelloReq struct {
	g.Meta `path:"/demo/hello"`
	Name   string `json:"name" v:"required"`
}

// demoHelloRes is the bound response type for demoController.Hello.
type demoHelloRes struct {
	Message string   `json:"message"`
	Trace   []string `json:"trace"`
}

// Hello returns a greeting and the middleware trace from runtime context.
func (c *demoController) Hello(ctx context.Context, req *demoHelloReq) (*demoHelloRes, error) {
	trace, _ := ctx.Value(traceKey{}).([]string)
	return &demoHelloRes{
		Message: "hello " + req.Name,
		Trace:   append([]string(nil), trace...),
	}, nil
}

// TestJsonrpcServerGroupMiddlewareBind verifies group middleware and controller binding.
func TestJsonrpcServerGroupMiddlewareBind(t *testing.T) {
	s := newDemoServer()

	rpcReq := jsonrpc.NewJsonrpcrequest()
	rpcReq.Method = "/api/demo/hello"
	rpcReq.Params = map[string]any{"name": "towgo"}
	rpcReq.Id = "1"

	rpcResp := s.Exec(context.Background(), rpcReq)
	if rpcResp.Error.Code != 200 {
		t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
	}

	var result demoHelloRes
	mustDecodeResult(t, rpcResp.Result, &result)

	if result.Message != "hello towgo" {
		t.Fatalf("unexpected message: %q", result.Message)
	}
	if got, want := strings.Join(result.Trace, ","), "server,group"; got != want {
		t.Fatalf("unexpected middleware trace: %q", got)
	}
}

// TestJsonrpcServerServeHTTP verifies JSON-RPC execution through net/http.
func TestJsonrpcServerServeHTTP(t *testing.T) {
	s := newDemoServer()
	body := `{"jsonrpc":"2.0","method":"/api/demo/hello","params":{"name":"http"},"id":"2"}`

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(body))
	s.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected http status: %d", rr.Code)
	}

	var rpcResp struct {
		Error  jsonrpc.Error `json:"error"`
		Result demoHelloRes  `json:"result"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &rpcResp); err != nil {
		t.Fatalf("response is not valid json: %s", rr.Body.String())
	}
	if rpcResp.Error.Code != 200 {
		t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
	}
	if rpcResp.Result.Message != "hello http" {
		t.Fatalf("unexpected response body: %s", rr.Body.String())
	}
}

// TestJsonrpcServerConnectionCompatibility verifies legacy connection-style handlers.
func TestJsonrpcServerConnectionCompatibility(t *testing.T) {
	s := jsonrpc.NewJsonrpcServer()
	s.MiddlewareConnection(func(conn jsonrpc.JsonRpcConnection) {
		conn.SetValue("trace", []string{"connection"})
		conn.Next()
	})
	s.HandleConnection("/legacy", func(conn jsonrpc.JsonRpcConnection) {
		trace, _ := conn.GetValue("trace")
		conn.WriteResult(map[string]any{
			"linkType": conn.LinkType(),
			"trace":    trace,
		})
	})

	rpcReq := jsonrpc.NewJsonrpcrequest()
	rpcReq.Method = "/legacy"
	rpcReq.Id = "3"
	conn := jsonrpc.NewLocalRpcConnection(rpcReq, nil)

	rpcResp := s.ExecConnection(conn)
	if rpcResp.Error.Code != 200 {
		t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
	}

	var result struct {
		LinkType string   `json:"linkType"`
		Trace    []string `json:"trace"`
	}
	mustDecodeResult(t, rpcResp.Result, &result)

	if result.LinkType != "local" {
		t.Fatalf("unexpected link type: %q", result.LinkType)
	}
	if got, want := strings.Join(result.Trace, ","), "connection"; got != want {
		t.Fatalf("unexpected trace: %q", got)
	}
}

// TestJsonrpcServerContextStaysOnRequest verifies runtime context is not on Jsonrpcrequest.
func TestJsonrpcServerContextStaysOnRequest(t *testing.T) {
	s := jsonrpc.NewJsonrpcServer()
	s.Middleware(func(r *jsonrpc.Request) {
		r.SetCtx(context.WithValue(r.Context(), ctxIsolationKey{}, "request"))
		r.Next()
	})
	s.Handle("/ctx", func(r *jsonrpc.Request) {
		_, hasContextMethod := reflect.TypeOf(r.GetRpcRequest()).MethodByName("Context")
		r.SetResult(map[string]any{
			"requestCtx":              r.Context().Value(ctxIsolationKey{}),
			"jsonrpcHasContextMethod": hasContextMethod,
		})
	})

	rpcReq := jsonrpc.NewJsonrpcrequest()
	rpcReq.Method = "/ctx"
	rpcResp := s.Exec(context.WithValue(context.Background(), ctxIsolationKey{}, "base"), rpcReq)
	if rpcResp.Error.Code != 200 {
		t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
	}

	var result struct {
		RequestCtx              string `json:"requestCtx"`
		JsonrpcHasContextMethod bool   `json:"jsonrpcHasContextMethod"`
	}
	mustDecodeResult(t, rpcResp.Result, &result)

	if result.RequestCtx != "request" {
		t.Fatalf("unexpected request ctx value: %q", result.RequestCtx)
	}
	if result.JsonrpcHasContextMethod {
		t.Fatal("jsonrpc request should not expose execution context")
	}
	if _, hasContextMethod := reflect.TypeOf(rpcReq).MethodByName("Context"); hasContextMethod {
		t.Fatal("Jsonrpcrequest should not expose Context method")
	}
}

// newDemoServer builds the shared GoFrame-style demo JSON-RPC server.
func newDemoServer() *jsonrpc.JsonrpcServer {
	s := jsonrpc.NewJsonrpcServer()
	s.Middleware(func(r *jsonrpc.Request) {
		r.SetCtx(context.WithValue(r.Context(), traceKey{}, []string{"server"}))
		r.Next()
	})
	s.Group("/api", func(group *jsonrpc.RouterGroup) {
		group.Middleware(func(r *jsonrpc.Request) {
			trace, _ := r.Context().Value(traceKey{}).([]string)
			trace = append(append([]string(nil), trace...), "group")
			r.SetCtx(context.WithValue(r.Context(), traceKey{}, trace))
			r.Next()
		})
		group.Bind(&demoController{})
	})
	return s
}

// mustDecodeResult decodes a JSON-like result into a typed test destination.
func mustDecodeResult(t *testing.T, value any, dest any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		t.Fatal(err)
	}
}

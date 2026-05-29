package towgo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/towgo/towgo/v2/towgo"
)

type ghttpJsonRpcDemoController struct {
}

type ghttpJsonRpcDemoReq struct {
	g.Meta `path:"/demo/hello"`
	Name   string `json:"name"`
}

type ghttpJsonRpcDemoRes struct {
	Message string `json:"message"`
}

func (c *ghttpJsonRpcDemoController) Hello(ctx context.Context, req *ghttpJsonRpcDemoReq) (*ghttpJsonRpcDemoRes, error) {
	return &ghttpJsonRpcDemoRes{
		Message: "hello " + req.Name,
	}, nil
}

func TestGhttpJsonRpcMiddleware(t *testing.T) {
	towgo.ResetForTest()
	towgo.BindObject("/demo", &ghttpJsonRpcDemoController{})

	s := g.Server(guid.S())
	s.SetPort(0)
	s.SetDumpRouterMap(false)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(towgo.JsonRpcRequest, towgo.JsonRpcResponse)
		group.ALL("/jsonrpc", towgo.GhttpMiddlewareHandler)
	})

	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	defer s.Shutdown()
	time.Sleep(100 * time.Millisecond)

	body := `{"jsonrpc":"2.0","method":"/demo/hello","params":{"name":"towgo"},"id":"1"}`
	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/jsonrpc", s.GetListenedPort()),
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	var rpcResp struct {
		Error  towgo.Error `json:"error"`
		Result struct {
			Message string `json:"message"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		t.Fatalf("response is not valid json: %s", respBytes)
	}
	if rpcResp.Error.Code != 200 {
		t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
	}
	if rpcResp.Result.Message != "hello towgo" {
		t.Fatalf("unexpected result: %s", respBytes)
	}
}

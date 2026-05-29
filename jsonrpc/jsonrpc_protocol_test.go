package jsonrpc_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/towgo/towgo/v2/jsonrpc"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"golang.org/x/net/websocket"
)

// protocolResult is the shared response shape for transport protocol tests.
type protocolResult struct {
	Protocol string `json:"protocol"`
	Marker   string `json:"marker"`
}

// TestJsonrpcServerProtocolConnections verifies local, HTTP, TCP, and websocket connection adapters.
func TestJsonrpcServerProtocolConnections(t *testing.T) {
	server := newProtocolServer()
	body := jsonrpcBody("/protocol/link", "1", nil)

	tests := []struct {
		name     string
		want     string
		newConn  func(t *testing.T) jsonrpc.JsonRpcConnection
		teardown func()
	}{
		{
			name: "local",
			want: "local",
			newConn: func(t *testing.T) jsonrpc.JsonRpcConnection {
				req := jsonrpc.NewJsonrpcrequest()
				req.Method = "/protocol/link"
				req.Id = "local"
				return jsonrpc.NewLocalRpcConnection(req, nil)
			},
		},
		{
			name: "http",
			want: "http",
			newConn: func(t *testing.T) jsonrpc.JsonRpcConnection {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(body))
				conn := jsonrpc.NewHttpRpcConnection(rr, req)
				conn.WithContext(req.Context())
				return conn
			},
		},
		{
			name: "tcp",
			want: "tcp",
			newConn: func(t *testing.T) jsonrpc.JsonRpcConnection {
				serverConn, clientConn := net.Pipe()
				t.Cleanup(func() {
					_ = serverConn.Close()
					_ = clientConn.Close()
				})
				return jsonrpc.NewTcpRpcConnection(serverConn, body)
			},
		},
		{
			name: "websocket",
			want: "websocket",
			newConn: func(t *testing.T) jsonrpc.JsonRpcConnection {
				conn := jsonrpc.NewWebSocketRpcConnection(nil)
				conn.AnalysisByString(body)
				return conn
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcResp := server.ExecConnection(tt.newConn(t))
			if rpcResp.Error.Code != 200 {
				t.Fatalf("unexpected rpc error: %+v", rpcResp.Error)
			}

			var result protocolResult
			mustDecodeResult(t, rpcResp.Result, &result)

			if result.Protocol != tt.want {
				t.Fatalf("unexpected protocol: got %q want %q", result.Protocol, tt.want)
			}
			if result.Marker != "middleware" {
				t.Fatalf("middleware did not run: %+v", result)
			}
		})
	}
}

// TestJsonrpcServerHTTPServerCommunication verifies net/http server-to-client JSON-RPC exchange.
func TestJsonrpcServerHTTPServerCommunication(t *testing.T) {
	server := httptest.NewServer(newProtocolServer())
	defer server.Close()

	resp, err := http.Post(
		server.URL,
		"application/json",
		strings.NewReader(jsonrpcBody("/protocol/link", "http-server", nil)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Error  jsonrpc.Error  `json:"error"`
		Result protocolResult `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatal(err)
	}
	if rpcResp.Error.Code != 200 || rpcResp.Result.Protocol != "http" {
		t.Fatalf("unexpected response: %+v", rpcResp)
	}
}

// TestJsonrpcServerTCPServerCommunication verifies newline-framed TCP JSON-RPC exchange.
func TestJsonrpcServerTCPServerCommunication(t *testing.T) {
	server := newProtocolServer()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()

		raw, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			serverDone <- err
			return
		}
		rpcConn := jsonrpc.NewTcpRpcConnection(conn, strings.TrimSpace(raw))
		server.ServeConnection(rpcConn)
		serverDone <- nil
	}()

	client, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := client.Write([]byte(jsonrpcBody("/protocol/link", "tcp-server", nil) + "\n")); err != nil {
		t.Fatal(err)
	}
	raw, err := bufio.NewReader(client).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if err := <-serverDone; err != nil {
		t.Fatal(err)
	}

	var rpcResp struct {
		Error  jsonrpc.Error  `json:"error"`
		Result protocolResult `json:"result"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &rpcResp); err != nil {
		t.Fatalf("response is not valid json: %s", raw)
	}
	if rpcResp.Error.Code != 200 || rpcResp.Result.Protocol != "tcp" {
		t.Fatalf("unexpected response: %s", raw)
	}
}

// TestJsonrpcServerServeConnectionWritesHTTPAndTCP verifies response writing for connection adapters.
func TestJsonrpcServerServeConnectionWritesHTTPAndTCP(t *testing.T) {
	server := newProtocolServer()
	body := jsonrpcBody("/protocol/link", "2", nil)

	t.Run("http", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/jsonrpc", strings.NewReader(body))
		conn := jsonrpc.NewHttpRpcConnection(rr, req)
		conn.WithContext(req.Context())

		server.ServeConnection(conn)

		var rpcResp struct {
			Error  jsonrpc.Error  `json:"error"`
			Result protocolResult `json:"result"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &rpcResp); err != nil {
			t.Fatalf("response is not valid json: %s", rr.Body.String())
		}
		if rpcResp.Error.Code != 200 || rpcResp.Result.Protocol != "http" {
			t.Fatalf("unexpected response: %s", rr.Body.String())
		}
	})

	t.Run("tcp", func(t *testing.T) {
		serverConn, clientConn := net.Pipe()
		defer serverConn.Close()
		defer clientConn.Close()

		conn := jsonrpc.NewTcpRpcConnection(serverConn, body)
		done := make(chan struct{})
		go func() {
			defer close(done)
			server.ServeConnection(conn)
		}()

		_ = clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		raw, err := bufio.NewReader(clientConn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		<-done

		var rpcResp struct {
			Error  jsonrpc.Error  `json:"error"`
			Result protocolResult `json:"result"`
		}
		if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &rpcResp); err != nil {
			t.Fatalf("response is not valid json: %s", raw)
		}
		if rpcResp.Error.Code != 200 || rpcResp.Result.Protocol != "tcp" {
			t.Fatalf("unexpected response: %s", raw)
		}
	})
}

// TestJsonrpcServerServeConnectionWritesWebSocket verifies websocket response writing.
func TestJsonrpcServerServeConnectionWritesWebSocket(t *testing.T) {
	server := newProtocolServer()
	body := jsonrpcBody("/protocol/link", "3", nil)

	wsServer := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			return
		}

		conn := jsonrpc.NewWebSocketRpcConnection(ws)
		conn.AnalysisByString(message)
		server.ServeConnection(conn)
	}))
	defer wsServer.Close()

	wsURL := "ws" + strings.TrimPrefix(wsServer.URL, "http")
	ws, err := websocket.Dial(wsURL, "", wsServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	if err := websocket.Message.Send(ws, body); err != nil {
		t.Fatal(err)
	}
	var raw string
	if err := websocket.Message.Receive(ws, &raw); err != nil {
		t.Fatal(err)
	}

	var rpcResp struct {
		Error  jsonrpc.Error  `json:"error"`
		Result protocolResult `json:"result"`
	}
	if err := json.Unmarshal([]byte(raw), &rpcResp); err != nil {
		t.Fatalf("response is not valid json: %s", raw)
	}
	if rpcResp.Error.Code != 200 || rpcResp.Result.Protocol != "websocket" {
		t.Fatalf("unexpected response: %s", raw)
	}
}

// TestJsonrpcServerGhttpRoute verifies GoFrame ghttp route integration.
func TestJsonrpcServerGhttpRoute(t *testing.T) {
	rpcServer := newDemoServer()
	httpServer := g.Server(guid.S())
	httpServer.SetPort(0)
	httpServer.SetDumpRouterMap(false)
	httpServer.Group("/", func(group *ghttp.RouterGroup) {
		group.ALL("/jsonrpc", rpcServer.GhttpHandler)
	})

	if err := httpServer.Start(); err != nil {
		t.Fatal(err)
	}
	defer httpServer.Shutdown()
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/jsonrpc", httpServer.GetListenedPort()),
		"application/json",
		strings.NewReader(jsonrpcBody("/api/demo/hello", "4", map[string]any{"name": "ghttp"})),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var rpcResp struct {
		Error  jsonrpc.Error `json:"error"`
		Result demoHelloRes  `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		t.Fatal(err)
	}
	if rpcResp.Error.Code != 200 || rpcResp.Result.Message != "hello ghttp" {
		t.Fatalf("unexpected response: %+v", rpcResp)
	}
}

// newProtocolServer builds a server that reports the active transport type.
func newProtocolServer() *jsonrpc.JsonrpcServer {
	server := jsonrpc.NewJsonrpcServer()
	server.MiddlewareConnection(func(conn jsonrpc.JsonRpcConnection) {
		conn.SetValue("marker", "middleware")
		conn.Next()
	})
	server.HandleConnection("/protocol/link", func(conn jsonrpc.JsonRpcConnection) {
		marker, _ := conn.GetValue("marker")
		conn.WriteResult(protocolResult{
			Protocol: conn.LinkType(),
			Marker:   fmt.Sprint(marker),
		})
	})
	return server
}

// jsonrpcBody creates a compact JSON-RPC request body for protocol tests.
func jsonrpcBody(method, id string, params map[string]any) string {
	if params == nil {
		params = map[string]any{}
	}
	data, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      id,
	})
	return string(data)
}

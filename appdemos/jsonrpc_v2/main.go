package main

import (
	"bufio"
	"context"
	"github.com/towgo/towgo/v2/jsonrpc"
	"log"
	"net"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"golang.org/x/net/websocket"
)

type protocolKey struct{}

type helloController struct{}

type helloReq struct {
	g.Meta `path:"/hello"`
	Name   string `json:"name" p:"name" d:"towgo"`
}

type helloRes struct {
	Message  string `json:"message"`
	Protocol string `json:"protocol"`
}

func (c *helloController) Hello(ctx context.Context, req *helloReq) (*helloRes, error) {
	protocol, _ := ctx.Value(protocolKey{}).(string)
	if protocol == "" {
		protocol = "http"
	}
	return &helloRes{
		Message:  "hello " + req.Name,
		Protocol: protocol,
	}, nil
}

func main() {
	rpcServer := newJsonrpcServer()

	go serveTCP(rpcServer, "127.0.0.1:8298")

	s := g.Server()
	s.SetPort(8199)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Bind(&helloController{})
		group.ALL("/jsonrpc", rpcServer.GhttpHandler)
		group.ALL("/ws", func(r *ghttp.Request) {
			websocket.Handler(func(ws *websocket.Conn) {
				serveWebSocket(rpcServer, ws)
			}).ServeHTTP(r.Response.Writer, r.Request)
		})
	})
	s.Run()
}

func newJsonrpcServer() *jsonrpc.JsonrpcServer {
	rpcServer := jsonrpc.NewJsonrpcServer()
	rpcServer.Group("/", func(group *jsonrpc.RouterGroup) {
		group.Middleware(func(r *jsonrpc.Request) {
			protocol := "jsonrpc"
			if conn := r.Connection(); conn != nil {
				protocol = conn.LinkType()
			}
			r.SetCtx(context.WithValue(r.Context(), protocolKey{}, protocol))
			r.Next()
		})
		group.Bind(&helloController{})
	})
	return rpcServer
}

func serveTCP(rpcServer *jsonrpc.JsonrpcServer, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("tcp listen failed: %v", err)
		return
	}
	log.Printf("jsonrpc tcp listening on %s", listener.Addr().String())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("tcp accept failed: %v", err)
			return
		}
		go serveTCPConn(rpcServer, conn)
	}
}

func serveTCPConn(rpcServer *jsonrpc.JsonrpcServer, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		rpcConn := jsonrpc.NewTcpRpcConnection(conn, strings.TrimSpace(message))
		rpcServer.ServeConnection(rpcConn)
	}
}

func serveWebSocket(rpcServer *jsonrpc.JsonrpcServer, ws *websocket.Conn) {
	for {
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			return
		}
		rpcConn := jsonrpc.NewWebSocketRpcConnection(ws)
		rpcConn.AnalysisByString(message)
		rpcServer.ServeConnection(rpcConn)
	}
}

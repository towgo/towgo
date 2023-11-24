package socks

import (
	"io"
	"log"
	"net"
	"strconv"

	"github.com/towgo/towgo/lib/aes"
)

type SocksServer struct {
	listenHost string
}

func (s *SocksServer) serverHandleSocks5(conn net.Conn) {
	defer func() {
		err := recover()
		if err != nil {
			log.Print(err)
		}
		conn.Close()
	}()

	//用加密通道进行通讯
	aesConn := aes.NewIOWrapper(conn)

	// 读取客户端的请求
	buf := make([]byte, 256)
	_, err := aesConn.Read(buf)
	if err != nil {
		log.Println("Failed to read client request:", err)
		return
	}

	// 解析客户端请求
	if buf[0] != 0x05 {
		log.Println("Invalid SOCKS5 request")
		return
	}

	// 回应客户端，告知支持无需身份验证
	aesConn.Write([]byte{0x05, 0x00})

	// 读取客户端的连接请求
	n, err := aesConn.Read(buf)
	if err != nil {
		log.Println("Failed to read client connection request:", err)
		return
	}

	// 解析连接请求
	if buf[0] != 0x05 || buf[1] != 0x01 || buf[2] != 0x00 {
		log.Println("Invalid SOCKS5 connection request")
		return
	}

	// 解析目标地址
	var targetAddr string
	switch buf[3] {
	case 0x01: // IPv4地址
		targetAddr = net.IPv4(buf[4], buf[5], buf[6], buf[7]).String()
	case 0x03: // 域名
		targetAddr = string(buf[5 : n-2])
	case 0x04: // IPv6地址
		targetAddr = net.IP(buf[4 : 4+net.IPv6len]).String()
	default:
		log.Println("Invalid address type")
		return
	}

	// 解析目标端口
	targetPort := strconv.Itoa(int(buf[n-2])<<8 | int(buf[n-1]))

	// 建立与目标服务器的连接
	targetConn, err := net.Dial("tcp", net.JoinHostPort(targetAddr, targetPort))
	if err != nil {
		log.Println("Failed to connect to target:", err)
		return
	}
	defer targetConn.Close()

	// 回应客户端，告知连接已建立
	aesConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// 进行转发
	go func(aesConn io.ReadWriter, conn net.Conn) {
		io.Copy(aesConn, conn)
		conn.Close()
	}(aesConn, targetConn)

	io.Copy(targetConn, aesConn)

}

func (s *SocksServer) ListenAndSrv() {
	// 监听本地端口
	listener, err := net.Listen("tcp", s.listenHost)
	if err != nil {
		log.Fatal("Failed to start SOCKS5 proxy:", err)
	}
	defer listener.Close()

	log.Println("Aes SOCKS5 proxy server started on " + s.listenHost)

	// 接受客户端连接并处理
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept client connection:", err)
			continue
		}

		go s.serverHandleSocks5(conn)
	}
}

func NewServer(listenHost string) *SocksServer {
	return &SocksServer{
		listenHost: listenHost,
	}
}

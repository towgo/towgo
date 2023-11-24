package socks

import (
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/towgo/towgo/lib/aes"
)

type ClientConnection struct {
	net.Conn
	targetBuf []byte
}

// 直连模式
func (s *SocksClient) directMode(conn *ClientConnection) error {

	// 读取客户端的请求
	buf := make([]byte, 256)
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}

	// 解析客户端请求
	if buf[0] != 0x05 {
		return errors.New("invalid SOCKS5 request")
	}

	// 回应客户端，告知支持无需身份验证
	conn.Write([]byte{0x05, 0x00})

	// 读取客户端的连接请求
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	conn.targetBuf = buf[0:n]

	// 解析连接请求
	if buf[0] != 0x05 || buf[1] != 0x01 || buf[2] != 0x00 {
		return errors.New("invalid SOCKS5 connection request")
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
		return errors.New("invalid address type")
	}

	// 解析目标端口
	targetPort := strconv.Itoa(int(buf[n-2])<<8 | int(buf[n-1]))

	// 建立与目标服务器的连接
	targetConn, err := net.DialTimeout("tcp", net.JoinHostPort(targetAddr, targetPort), time.Millisecond*350)
	if err != nil {
		return err
	}
	defer targetConn.Close()

	// 回应客户端，告知连接已建立
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// 进行转发
	go func(sourceConn net.Conn, targetConn net.Conn) {
		io.Copy(sourceConn, targetConn)
		sourceConn.Close()
	}(conn, targetConn)

	io.Copy(targetConn, conn)

	return nil
}

func (s *SocksClient) clientHandleSocks5(conn *ClientConnection) {
	defer func() {
		err := recover()
		if err != nil {
			log.Print(err)
		}
		conn.Close()
	}()

	//智能分路模式
	err := s.directMode(conn)
	if err == nil {

		//没有错误返回说明已经选择了本地直连模式
		return
	}

	//代理模式

	//建立加密通道
	// 建立与目标服务器的连接
	serverConn, err := net.Dial("tcp", s.serverHost)
	if err != nil {
		log.Println("Failed to connect to target:", err)
		return
	}
	defer serverConn.Close()

	//与服务端建立aes加密通道
	aesConn := aes.NewIOWrapper(serverConn)

	aesConn.Write([]byte{0x05})

	buf := make([]byte, 256)
	aesConn.Read(buf)
	if buf[0] != 0x05 || buf[1] != 0x00 {
		log.Print("错误的socks请求")
		return
	}

	_, err = aesConn.Write(conn.targetBuf)
	if err != nil {
		log.Print(err.Error())
		return
	}

	//双向数据拷贝,建立通道
	go func(aesConn io.ReadWriter, conn net.Conn) {
		io.Copy(aesConn, conn)
		conn.Close()
	}(aesConn, conn)

	io.Copy(conn, aesConn)

}

type SocksClient struct {
	localListenHost string
	serverHost      string
}

func (s *SocksClient) ListenAndSrv() {
	// 监听本地端口
	listener, err := net.Listen("tcp", s.localListenHost)
	if err != nil {
		log.Fatal("Failed to start SOCKS5 proxy:", err)
	}
	defer listener.Close()

	log.Println("SOCKS5 proxy started on " + s.localListenHost)

	// 接受客户端连接并处理
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept client connection:", err)
			continue
		}
		connection := &ClientConnection{
			Conn: conn,
		}

		go s.clientHandleSocks5(connection)
	}
}

func NewClient(localListenHost string, serverHost string) *SocksClient {
	return &SocksClient{
		localListenHost: localListenHost,
		serverHost:      serverHost,
	}
}

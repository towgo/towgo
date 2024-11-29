/*
JSON-RPC2.0 over serial for golang
by:liangliangit
ver 1.0
*/

package towgo

type SerialConnConfig struct {
	Device string `json:"device"` //串口设备名称 eg: linux:/dev/ttyS1 or windows:COM1
}

type SerialConn struct {
	BUFFERLENGTH int64 //内存缓冲区长度 （字节） 1024000byte=100MB
	DATAEND      byte
}

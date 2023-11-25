package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/sigurn/crc16"
	"github.com/tarm/serial"
	"github.com/towgo/towgo/lib/system"
)

var devLock sync.Mutex
var basePath = system.GetPathOfProgram()

type Config struct {
	Serverport string `json:"serverport"`
}

type Response struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

var config Config

func main() {

	system.ScanConfigJson(basePath+"/config/config.json", &config)
	http.HandleFunc("/string_hex_tty", http_string_hex_tty)
	log.Print("0.0.0.0:" + config.Serverport + " 服务启动成功")
	err := http.ListenAndServe("0.0.0.0:"+config.Serverport, nil)
	if err != nil {
		log.Print(err.Error())
	}
}

func strToHex(str string) []byte {

	str = strings.Replace(str, " ", "", -1)
	// 将字符串转换为字节数组
	inputBytes, err := hex.DecodeString(str)

	if err != nil {
		fmt.Println("错误:", err)
		return nil
	}

	return inputBytes
}

func http_string_hex_tty(w http.ResponseWriter, r *http.Request) {
	var resp Response
	devLock.Lock()
	defer devLock.Unlock()
	// 配置串口参数
	config := &serial.Config{
		Name:     "/dev/tty.usbserial-2110", // 串口设备名，根据你的实际情况修改
		Baud:     9600,                      // 波特率
		Size:     0,
		Parity:   0,
		StopBits: 1,
	}
	config.ReadTimeout = time.Second * 1

	config.Name = r.URL.Query().Get("dev")

	if config.Name == "" {
		resp.Code = 500
		resp.Msg = "串口设备不能为空"
		http_write_json(resp, w)
		return
	}

	// 设置要分配的权限，例如0666表示读写权限
	permissions := "0666"

	// 使用chmod命令修改权限
	cmd := exec.Command("chmod", permissions, config.Name)

	// 执行命令
	err := cmd.Run()
	if err != nil {
		log.Print("无法修改权限：%v", err)
		resp.Code = 500
		resp.Msg = "无法修改权限："
		http_write_json(resp, w)
		return
	}

	str := r.URL.Query().Get("str")

	if str == "" {
		resp.Code = 500
		resp.Msg = "发送数据不能为空"
		http_write_json(resp, w)
		return
	}

	h := strToHex(str)
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Print(err)
		resp.Code = 500
		resp.Msg = err.Error()
		http_write_json(resp, w)
		return
	}

	defer port.Close()

	_, err = port.Write(h)
	if err != nil {
		log.Print(err)
		resp.Code = 500
		resp.Msg = err.Error()
		http_write_json(resp, w)
		return
	}

	var data string
	var dataByte []byte

	timeout := time.NewTimer(time.Second * 2)

	for {
		select {
		case <-timeout.C:
			log.Print(err)
			resp.Code = 500
			resp.Msg = "读取超时"
			http_write_json(resp, w)
			return

		default:
			// 从串口读取数据
			buf := make([]byte, 1024)
			//log.Print("读取数据")
			n, err := port.Read(buf)
			if err != nil {
				log.Print(err)
				resp.Code = 200
				resp.Data = hex.EncodeToString(dataByte)
				http_write_json(resp, w)
				return
			}

			dataByte = append(dataByte, buf[:n]...)

			crcValue := calculateCRC16(dataByte[0 : len(dataByte)-2])

			// 创建一个长度为2的子节切片
			bytes := make([]byte, 2)

			// 使用binary.LittleEndian或binary.BigEndian将uint16值转换为子节
			binary.LittleEndian.PutUint16(bytes, crcValue)

			if string(bytes) == string(dataByte[len(dataByte)-2:]) {
				data = hex.EncodeToString(dataByte)
				resp.Code = 200
				resp.Data = data
				http_write_json(resp, w)
				return
			}
		}

	}

}

func http_write_json(obj any, w http.ResponseWriter) {
	b, _ := json.Marshal(obj)
	w.Write(b)
}

func calculateCRC16(data []byte) uint16 {
	table := crc16.MakeTable(crc16.CRC16_MODBUS)

	crc := crc16.Checksum(data, table)

	return crc
	//return swapBytes(crc)
}

func swapBytes(value uint16) uint16 {
	// 将高8位和低8位进行交换
	return (value << 8) | (value >> 8)
}

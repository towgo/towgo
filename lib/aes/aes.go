package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
)

var key = []byte("KHGSI69YBWGS0TWX")
var iv = []byte("3010201735544643")

const (
	DATAEND byte = '\n' //数据尾帧标识符 (防止粘包)
)

type IO struct {
	readBufSize       int
	maxDataLength     int
	readBufEncodeData []byte
	readBufData       []byte
	rw                io.ReadWriteCloser
}

func (i *IO) Close() error {
	return i.rw.Close()
}

func (i *IO) Write(p []byte) (n int, err error) {
	b, err := AesEncrypt(p)
	if err != nil {
		return 0, err
	}

	b = append(b, DATAEND)

	_, err = i.rw.Write(b)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (i *IO) cutRead(p []byte) (n int, err error) {
	readBufDataLen := len(i.readBufData)
	pLen := len(p)

	n = copy(p, i.readBufData)

	if readBufDataLen > pLen {
		i.readBufData = i.readBufData[n-1:]
	} else {
		i.readBufData = []byte{}
	}

	return
}

func (i *IO) Read(p []byte) (n int, err error) {

	if len(i.readBufData) > 0 {
		return i.cutRead(p)
	}

	var readDataBuf []byte
	if len(i.readBufEncodeData) > 0 {
		readDataBuf = append(readDataBuf, i.readBufEncodeData...)
		i.readBufEncodeData = []byte{}
	}
	for {

		readDataBufLen := len(readDataBuf)

		if readDataBufLen > i.maxDataLength {
			return 0, errors.New("超过加密允许的最大数据长度")
		}

		//读取数据
		readBuf := make([]byte, i.readBufSize)
		rn, err := i.rw.Read(readBuf)
		if err != nil {
			return 0, err
		}

		rbf := readBuf[:rn]

		//将读到的数据追加到缓冲区
		readDataBuf = append(readDataBuf, rbf...)

		//等待尾帧数据
		if !bytes.Contains(readDataBuf, []byte{DATAEND}) {
			continue
		}

		//有尾帧数据,数据完整,可以解密

		//拆包
		spBuf := bytes.Split(readDataBuf, []byte{DATAEND})

		for k, v := range spBuf {
			if len(v) == 0 {
				continue
			}

			if len(spBuf)-1 == k {
				if v[len(v)-1] == DATAEND {
					b, err := AesDecrypt(string(v[0 : len(v)-1]))
					if err != nil {
						log.Print(err.Error())
						return 0, err
					}
					i.readBufData = append(i.readBufData, b...)

				} else {
					i.readBufEncodeData = v
				}
			} else {
				b, err := AesDecrypt(string(v))
				if err != nil {
					log.Print(err.Error())
					return 0, err
				}
				i.readBufData = append(i.readBufData, b...)
			}

		}

		return i.cutRead(p)
	}
}

func NewIOWrapper(rw io.ReadWriteCloser) *IO {
	return &IO{
		readBufSize:   1,       //读缓冲区1Mb
		maxDataLength: 1024000, //加密数据允许的最大长度100Mb
		rw:            rw,
	}
}

func SetKey(s string) {
	key = []byte(s)
}
func SetIv(s string) {
	iv = []byte(s)
}

func AesEncryptStruct(s any) (encrypt []byte, err error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return AesEncrypt(b)
}

func AesDecryptStruct(destStruct any, text string) error {
	b, err := AesDecrypt(text)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, destStruct)
}

func AesEncrypt(text []byte) (encrypt []byte, err error) {
	//生成cipher.Block 数据块
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("错误 -" + err.Error())
		return nil, err
	}
	//填充内容，如果不足16位字符
	blockSize := block.BlockSize()
	originData := pad(text, blockSize)
	//加密方式
	blockMode := cipher.NewCBCEncrypter(block, iv)
	//加密，输出到[]byte数组
	crypted := make([]byte, len(originData))
	blockMode.CryptBlocks(crypted, originData)
	encrypt = []byte(base64.StdEncoding.EncodeToString(crypted))
	return
}

func pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AesDecrypt(text string) ([]byte, error) {
	decode_data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, err
	}

	//生成密码数据块cipher.Block
	block, _ := aes.NewCipher(key)
	//解密模式
	blockMode := cipher.NewCBCDecrypter(block, iv)
	//输出到[]byte数组
	origin_data := make([]byte, len(decode_data))
	blockMode.CryptBlocks(origin_data, decode_data)
	//去除填充,并返回
	return (unpad(origin_data)), nil
}

func unpad(ciphertext []byte) []byte {
	length := len(ciphertext)
	//去掉最后一次的padding
	unpadding := int(ciphertext[length-1])
	return ciphertext[:(length - unpadding)]
}

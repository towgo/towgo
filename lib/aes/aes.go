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
	canread           bool
	canwrite          bool
	canclose          bool
	iotype            string
	rwc               io.ReadWriteCloser
	r                 io.Reader
	w                 io.Writer
}

func (i *IO) Close() error {
	if !i.canclose {
		return errors.New("该对象不支持close方法")
	}
	return i.rwc.Close()
}

func (i *IO) Write(p []byte) (n int, err error) {
	if !i.canwrite {
		return 0, errors.New("该对象不支持write方法")
	}
	b, err := AesEncrypt(p)
	if err != nil {
		return 0, err
	}

	b = append(b, DATAEND)

	switch i.iotype {
	case "rwc":
		_, err = i.rwc.Write(b)
		if err != nil {
			return 0, err
		}
	case "w":
		_, err = i.w.Write(b)
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (i *IO) decode(readDataBuf []byte) error {
	//拆包
	spBuf := bytes.Split(readDataBuf, []byte{DATAEND})
	//log.Print("共找到", len(spBuf)-1, "组数据包")
	residualData := spBuf[len(spBuf)-1]
	intactEncodeData := spBuf[0 : len(spBuf)-1]
	if len(residualData) > 0 {
		i.readBufEncodeData = residualData
	} else {
		i.readBufEncodeData = nil
	}

	for _, v := range intactEncodeData {
		b, err := AesDecrypt(string(v))
		if err != nil {
			log.Print(err.Error())
			return err
		}
		i.readBufData = append(i.readBufData, b...)
	}
	return nil
}

func (i *IO) cutRead(p []byte) (n int, err error) {
	readBufDataLen := len(i.readBufData)
	pLen := len(p)
	n = copy(p, i.readBufData)
	if readBufDataLen > pLen {
		i.readBufData = i.readBufData[n:]
	} else {
		i.readBufData = []byte{}
	}
	return
}

func (i *IO) Read(p []byte) (n int, err error) {
	if !i.canread {
		return 0, errors.New("该对象不支持read方法")
	}
	if len(i.readBufData) > 0 {
		return i.cutRead(p)
	}

	var readDataBuf []byte
	if len(i.readBufEncodeData) > 0 {
		readDataBuf = append(readDataBuf, i.readBufEncodeData...)
		i.readBufEncodeData = nil
	}
	for {

		readDataBufLen := len(readDataBuf)

		if readDataBufLen > i.maxDataLength {
			log.Print("超过加密允许的最大数据长度")
			return 0, errors.New("超过加密允许的最大数据长度")
		}

		//读取数据
		readBuf := make([]byte, i.readBufSize)

		var rn int
		var err error

		switch i.iotype {
		case "rwc":
			rn, err = i.rwc.Read(readBuf)
		case "r":
			rn, err = i.r.Read(readBuf)
		}

		//log.Print("读到了", rn, "条数据")

		if rn > 0 {
			//将读到的数据追加到缓冲区
			readDataBuf = append(readDataBuf, readBuf[:rn]...)
		}

		if err != nil {
			if rn > 0 {
				i.decode(readDataBuf)
				return i.cutRead(p)
			}
			return 0, err
		}

		//等待尾帧数据
		if !bytes.Contains(readDataBuf, []byte{DATAEND}) {
			log.Print("等待尾帧")
			continue
		}

		//有尾帧数据,数据完整,可以解密
		log.Print("有尾帧数据,数据完整,可以解密")
		i.decode(readDataBuf)
		return i.cutRead(p)
	}
}

func NewIOWrapper(rwc io.ReadWriteCloser) *IO {
	return &IO{
		canread:       true,
		canwrite:      true,
		canclose:      true,
		iotype:        "rwc",
		readBufSize:   102400,  //读缓冲区10MB
		maxDataLength: 1024000, //加密数据允许的最大长度100MB
		rwc:           rwc,
	}
}

func NewWriteWrapper(w io.Writer) *IO {
	return &IO{
		canwrite:      true,
		iotype:        "w",
		readBufSize:   102400,  //读缓冲区10MB
		maxDataLength: 1024000, //加密数据允许的最大长度100MB
		w:             w,
	}
}

func NewReadWrapper(r io.Reader) *IO {
	return &IO{
		canread:       true,
		iotype:        "r",
		readBufSize:   102400,  //读缓冲区10MB
		maxDataLength: 1024000, //加密数据允许的最大长度100MB
		r:             r,
	}
}

func SetKey(s string) {
	key = []byte(s)
}
func SetIv(s string) {

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
	defer func() {
		err := recover()
		if err != nil {
			log.Print(err)
		}
	}()
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
	defer func() {
		err := recover()
		if err != nil {
			log.Print(err)
		}
	}()
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

package system

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"
)

// 生成RSA私钥和公钥，保存到文件中
func GenerateRSAKey(bits int, outPath string, fileName string) {

	outPath = strings.TrimRight(outPath, GetPathSymbol()) + GetPathSymbol()

	if fileName != "" {
		fileName = fileName + "."
	}

	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Print(err)
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	// X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey) // PKCS1 和 9 是不一致的
	X509PrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	//使用pem格式对x509输出的内容进行编码
	//创建文件保存私钥
	privateFile, err := os.Create(outPath + fileName + "private.pem")
	if err != nil {
		log.Print(err)
	}
	defer privateFile.Close()
	//构建一个pem.Block结构体对象
	privateBlock := pem.Block{Type: "PRIVATE KEY", Bytes: X509PrivateKey}
	//将数据保存到文件
	pem.Encode(privateFile, &privateBlock)
	//保存公钥
	//获取公钥的数据
	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		log.Print(err)
	}
	//pem格式编码
	//创建用于保存公钥的文件
	publicFile, err := os.Create(outPath + fileName + "public.pem")
	if err != nil {
		log.Print(err)
	}
	defer publicFile.Close()
	//创建一个pem.Block结构体对象
	publicBlock := pem.Block{Type: "Public Key", Bytes: X509PublicKey}
	//保存到文件
	pem.Encode(publicFile, &publicBlock)
}

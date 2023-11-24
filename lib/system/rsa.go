package system

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

var publicKeyCache *rsa.PublicKey
var privateKeyCache *rsa.PrivateKey

func SetPublicKey(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
	}
	defer file.Close()
	//获取文件内容
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)

	publicKeyCache = getPublicKey(buf)

}

func SetPublicKeyByIO(fs io.Reader) {
	b, _ := io.ReadAll(fs)
	publicKeyCache = getPublicKey(b)
}

func getPublicKey(keyByte []byte) *rsa.PublicKey {
	block, _ := pem.Decode(keyByte)
	//X509解码
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	return publicKeyInterface.(*rsa.PublicKey)
}

func SetPrivateKey(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
	}
	defer file.Close()
	//获取文件内容
	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)
	privateKeyCache = getPrivateKey(buf)
}

func SetPrivateKeyByIO(fs io.Reader) {
	b, _ := io.ReadAll(fs)
	privateKeyCache = getPrivateKey(b)
}

func getPrivateKey(keyByte []byte) *rsa.PrivateKey {
	// pem解码
	block, _ := pem.Decode(keyByte)
	// X509解码
	privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	return privateKeyInterface.(*rsa.PrivateKey)
}

func DecryptPrivate(cipherText []byte) ([]byte, error) {

	if privateKeyCache == nil {
		return nil, errors.New("private key not set")
	}
	decode_data, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}

	var bytesDecrypt []byte

	keySize := privateKeyCache.Size()
	srcSize := len(decode_data)
	var offSet = 0
	var buffer = bytes.Buffer{}
	for offSet < srcSize {
		endIndex := offSet + keySize
		if endIndex > srcSize {
			endIndex = srcSize
		}
		bytesOnce, err := rsa.DecryptPKCS1v15(rand.Reader, privateKeyCache, decode_data[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}

	bytesDecrypt = buffer.Bytes()

	return bytesDecrypt, nil
}

func DecryptPublic(cipherText []byte) ([]byte, error) {

	if publicKeyCache == nil {
		return nil, errors.New("public key not set")
	}

	decode_data, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}

	var bytesDecrypt []byte

	keySize := publicKeyCache.Size()
	srcSize := len(decode_data)

	var offSet = 0
	var buffer = bytes.Buffer{}
	for offSet < srcSize {
		endIndex := offSet + keySize
		if endIndex > srcSize {
			endIndex = srcSize
		}
		bytesOnce, err := PublicDecrypt(publicKeyCache, decode_data[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}
	bytesDecrypt = buffer.Bytes()
	return bytesDecrypt, nil
}

// RsaEncryptBlock 公钥加密-分段
func EncryptPublic(src []byte) (bytesEncrypt []byte, err error) {
	//打开文件

	if publicKeyCache == nil {
		return nil, errors.New("public key not set")
	}

	keySize, srcSize := publicKeyCache.Size(), len(src)

	offSet, once := 0, keySize-11
	buffer := bytes.Buffer{}
	for offSet < srcSize {
		endIndex := offSet + once
		if endIndex > srcSize {
			endIndex = srcSize
		}
		// 加密一部分
		bytesOnce, err := rsa.EncryptPKCS1v15(rand.Reader, publicKeyCache, src[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}
	bytesEncrypt = []byte(base64.StdEncoding.EncodeToString(buffer.Bytes()))
	return
}

// RsaEncryptBlock 私钥加密-分段
func EncryptPrivate(src []byte) (bytesEncrypt []byte, err error) {

	if privateKeyCache == nil {
		return nil, errors.New("private key not set")
	}

	keySize, srcSize := privateKeyCache.Size(), len(src)

	offSet, once := 0, keySize-11
	buffer := bytes.Buffer{}
	for offSet < srcSize {
		endIndex := offSet + once
		if endIndex > srcSize {
			endIndex = srcSize
		}
		// 加密一部分
		bytesOnce, err := PrivateEncrypt(privateKeyCache, src[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}
	bytesEncrypt = []byte(base64.StdEncoding.EncodeToString(buffer.Bytes()))
	return
}

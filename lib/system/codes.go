package system

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"reflect"
	"regexp"
	"sync/atomic"
	"time"
)

var objectIdCounter uint32 = 0

var machineId = readMachineId()

type ObjectId string

func readMachineId() []byte {
	var sum [3]byte
	id := sum[:]
	hostname, err1 := os.Hostname()
	if err1 != nil {
		_, err2 := io.ReadFull(rand.Reader, id)
		if err2 != nil {
			log.Printf("cannot get hostname: %v; %v", err1, err2)
		}
		return id
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(id, hw.Sum(nil))
	//fmt.Println("readMachineId:" + string(id))
	return id
}

// GUID returns a new unique ObjectId.
// 4byte 时间，
// 3byte 机器ID
// 2byte pid
// 3byte 自增ID
func GetGUID() ObjectId {
	var b [12]byte
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b[:], uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&objectIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return ObjectId(b[:])
}

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}

// Hex returns a hex representation of the ObjectId.
// 返回16进制对应的字符串
func (id ObjectId) Hex() string {
	return hex.EncodeToString([]byte(id))
}

func RandChar(size int) string {
	char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	len64 := int64(len(char))
	var s bytes.Buffer
	for i := 0; i < size; i++ {
		in, _ := rand.Int(rand.Reader, big.NewInt(len64))
		s.WriteByte(char[in.Int64()])
	}
	return s.String()
}

func RandCharNumber(size int) string {
	char := "0123456789"
	len64 := int64(len(char))
	var s bytes.Buffer
	for i := 0; i < size; i++ {
		in, _ := rand.Int(rand.Reader, big.NewInt(len64))
		s.WriteByte(char[in.Int64()])
	}
	return s.String()
}

func RandCharCrypto(size int) string {
	char := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	len64 := int64(len(char))
	var s bytes.Buffer
	for i := 0; i < size; i++ {
		in, _ := rand.Int(rand.Reader, big.NewInt(len64))
		s.WriteByte(char[in.Int64()])
	}
	return s.String()
}

func RandChar16H(size int) string {
	char := "ABCDEF0123456789"
	len64 := int64(len(char))
	var s bytes.Buffer
	for i := 0; i < size; i++ {
		in, _ := rand.Int(rand.Reader, big.NewInt(len64))
		s.WriteByte(char[in.Int64()])
	}
	return s.String()
}

func FilteredSQLInject(to_match_str string) bool {
	//过滤 ‘
	//ORACLE 注解 --  /**/
	//关键字过滤 update ,delete
	// 正则的字符串, 不能用 " " 因为" "里面的内容会转义
	str := `(?:')|(?:--)|(/\\*(?:.|[\\n\\r])*?\\*/)|(\b(select|update|and|or|delete|insert|trancate|char|chr|into|substr|ascii|declare|exec|count|master|into|drop|execute)\b)`
	re, err := regexp.Compile(str)
	if err != nil {
		return false
	}
	return re.MatchString(to_match_str)
}

func MD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func SHA1(str string) string {
	has := sha1.Sum([]byte(str))
	return fmt.Sprintf("%x", has)
}

func SHA256(str string) string {
	has := sha256.New().Sum([]byte(str))
	return fmt.Sprintf("%x", has)
}

func MD5Any(object any) string {
	b, _ := json.Marshal(object)
	has := md5.Sum(b)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}

func MarshalWithEmptySlices(v any) ([]byte, error) {
	// 递归检测并替换nil切片为空切片
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return json.Marshal(v)
	}

	traverseStruct(&value) // 处理结构体字段
	return json.Marshal(v)
}

func traverseStruct(v *reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Slice && field.IsNil() {
			// 将nil切片初始化为空切片
			field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		} else if field.Kind() == reflect.Struct {
			traverseStruct(&field)
		}
	}
}

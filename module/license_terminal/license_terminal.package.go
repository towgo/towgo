package licenseterminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/towgo/towgo/module/dblog"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/jsonrpc"
	"github.com/towgo/towgo/lib/system"
)

var accessCode string
var expirationCallFuncs sync.Map //存放到期回调的注册函数
var activeCallFuncs sync.Map     //存放激活回调的函数
var isChecking bool              //自动检查是否运行
var pathSymbol string = system.GetPathSymbol()
var basePath string = system.GetPathOfProgram()
var serverPath string = filepath.Join(basePath, "config", "license.conf.json")
var productNumber string

func init() {

	jsonrpc.SetFunc("/license/list", func(rpcConn jsonrpc.JsonRpcConnection) {
		result := map[string]interface{}{}
		result["rows"] = GetLicense()
		rpcConn.WriteResult(result)
	})

	//获取激活码
	jsonrpc.SetFunc("/license/activecode/get", func(rpcConn jsonrpc.JsonRpcConnection) {

		err := activecodeRequestCheck(rpcConn)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
			rpcConn.Write()
			return
		}

		var requestParams struct {
			ProductNumber string `json:"product_number"`
			LicenseKey    string `json:"license_key"`
			AccessCode    string `json:"access_code"`
		}
		rpcConn.ReadParams(&requestParams)

		if requestParams.ProductNumber == "" {
			requestParams.ProductNumber = productNumber
		}

		var active ActiveRequestInfo

		if requestParams.AccessCode != "" {
			active.AccessCode = requestParams.AccessCode
		} else {
			active.AccessCode = GetAccessCode()
		}

		active.LicenseKey = requestParams.LicenseKey
		active.ProductNumber = requestParams.ProductNumber

		jb, _ := json.Marshal(active)
		b, err := system.EncryptPublic(jb)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
			rpcConn.Write()
			return
		}

		result := map[string]interface{}{}
		result["active_code"] = string(b)
		rpcConn.WriteResult(result)
	})

	//联机激活请求
	jsonrpc.SetFunc("/license/activecode/online/request", func(rpcConn jsonrpc.JsonRpcConnection) {
		var conf struct {
			LicenseServerUrl string
		}
		system.ScanConfigJson(serverPath, &conf)

		if conf.LicenseServerUrl == "" {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, "许可证服务器地址未配置")
			rpcConn.Write()
			return
		}

		err := activecodeRequestCheck(rpcConn)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
			rpcConn.Write()
			return
		}

		var requestParams struct {
			ProductNumber string `json:"product_number"`
			LicenseKey    string `json:"license_key"`
		}
		rpcConn.ReadParams(&requestParams)

		if requestParams.ProductNumber == "" {
			requestParams.ProductNumber = productNumber
		}

		var active ActiveRequestInfo
		active.AccessCode = GetAccessCode()
		active.LicenseKey = requestParams.LicenseKey
		active.ProductNumber = requestParams.ProductNumber

		jb, _ := json.Marshal(active)

		b, err := system.EncryptPublic(jb)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
			rpcConn.Write()
			return
		}

		request := jsonrpc.NewJsonrpcrequest()
		request.Method = "/license/getActiveCredential"
		var activeRequest struct {
			ActiveCode string `json:"active_code"`
		}
		activeRequest.ActiveCode = string(b)
		request.Params = activeRequest

		rpcClient := jsonrpc.NewHttpClient()
		rpcClient.ErrorFunc = func(err error) {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
		}
		rpcClient.Call(conf.LicenseServerUrl+"/jsonrpc", request, func(j jsonrpc.Jsonrpcresponse) {
			if j.Error.Code != 200 {
				rpcConn.GetRpcResponse().Error.Set(j.Error.Code, j.Error.Message)
				rpcConn.Write()
				return
			}
			var result struct {
				ActiveCredential string `json:"active_credential"`
			}
			j.ReadResult(&result)
			err = SaveLicense(filepath.Join(basePath, "license", requestParams.ProductNumber+".license"), result.ActiveCredential)
			if err != nil {
				rpcConn.GetRpcResponse().Error.Set(j.Error.Code, err.Error())
				rpcConn.Write()
				return
			}
			InstallReNewCheck()
			l, err := GetLicenseByName(requestParams.ProductNumber)
			if err != nil {
				rpcConn.GetRpcResponse().Error.Set(jsonrpc.JSONRPC_500_INTERNAL_SERVER_ERROR, err.Error())
				rpcConn.Write()
				return
			}
			rpcConn.WriteResult(l)
		})

	})

	jsonrpc.SetFunc("/license/active", func(rpcConn jsonrpc.JsonRpcConnection) {

		var params struct {
			ActiveCredential string `json:"active_credential"`
		}
		rpcConn.ReadParams(&params)

		l, err := DecodeLicense(params.ActiveCredential)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
		err = SaveLicense(filepath.Join(basePath, "license", l.OrderSerialNumber+".license"), params.ActiveCredential)
		if err != nil {
			rpcConn.GetRpcResponse().Error.Set(500, err.Error())
			rpcConn.Write()
			return
		}
		InstallReNewCheck()
		rpcConn.WriteResult("ok")
	})

	jsonrpc.SetFunc("/license/accesscode/get", func(rpcConn jsonrpc.JsonRpcConnection) {
		result := map[string]interface{}{}
		result["accesscode"] = GetAccessCode()
		rpcConn.WriteResult(result)
	})

	//获取授权码
	http.HandleFunc("/license/accesscode/get", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(GetAccessCode()))
	})
	//安装许可证
	http.HandleFunc("/license/install/renew", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("成功安装了" + strconv.FormatInt(InstallReNewCheck(), 10) + "个许可证"))
	})
	accessCode = GetAccessCode()
	go dblog.BatchInsert(
		dblog.NewOperateType(dblog.QUERY, "/license/list", "account_center", "查询许可证列表"),
		dblog.NewOperateType(dblog.QUERY, "/license/activecode/get", "account_center", "获取授权码"),
		dblog.NewOperateType(dblog.UPDATE, "/license/activecode/online/request", "account_center", "联机激活请求"),
		dblog.NewOperateType(dblog.UPLOAD, "/license/active", "account_center", "申请许可证"),
		dblog.NewOperateType(dblog.QUERY, "/license/accesscode/get", "account_center", "获取授权码"),
	)

	//go hardCheckExpiration()

	go autoCheckExpiration()
}

// 定时任务 检查许可证过期情况
func autoCheckExpiration() {

	if isChecking {
		return
	}
	for {
		isChecking = true
		time.Sleep(time.Second * 60)

		//注册函数检查
		licenses := GetLicense()
		expirationCallFuncs.Range(func(key, _ any) bool {

			//许可证查询
			for _, v := range licenses {
				if v.AccessCode != accessCode {
					log.Print("许可证与授权码不匹配")
					continue
				}

				//许可证存在
				if v.ProductNumber == key.(string) {
					//到期检查

					if v.Endtime > time.Now().UnixMilli() {
						//没有过期
						return true
					}

				}
			}

			//许可证不存在或过期
			notifyExpiration(key.(string))
			//继续迭代
			return true
		})
	}
}

func activecodeRequestCheck(rpcConn jsonrpc.JsonRpcConnection) error {
	var requestParams struct {
		ProductNumber string `json:"product_number"`
		LicenseKey    string `json:"license_key"`
	}

	rpcConn.ReadParams(&requestParams)
	if requestParams.LicenseKey == "" {
		return errors.New("license_key can not be null")
	}

	if requestParams.ProductNumber == "" {
		requestParams.ProductNumber = productNumber
	}

	licenseKeys := strings.Split(requestParams.LicenseKey, "-")
	if len(licenseKeys) != 4 {
		return errors.New("license_key error")
	}

	if strings.ToUpper(cehckSumString([]byte(requestParams.ProductNumber))) != licenseKeys[0] {
		return errors.New("license_key not support for product number")
	}

	check := licenseKeys[0] + "-" + licenseKeys[1] + "-" + licenseKeys[2]
	check = strings.ToUpper(check)
	crcStr := strings.ToUpper(cehckSumString([]byte(check)))
	if crcStr != licenseKeys[3] {
		return errors.New("license_key error")
	}
	return nil
}

// 到期通知回调程序（不传值代表广播，即代表所有的授权都过期了）
func notifyExpiration(ProductSerialNumber string) {
	if ProductSerialNumber == "" {
		expirationCallFuncs.Range(func(_, value any) bool {
			value.(func())()
			return true
		})
		return
	}

	value, ok := expirationCallFuncs.Load(ProductSerialNumber)
	if ok {
		value.(func())()
	}
}

func getActiveRequestInfoCode(licenseKey, productNumber string) string {
	var active ActiveRequestInfo
	active.AccessCode = GetAccessCode()
	active.LicenseKey = licenseKey
	active.ProductNumber = productNumber
	jb, _ := json.Marshal(active)
	b, err := system.EncryptPublic(jb)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// 安装许可证
func notifyReNew(ProductSerialNumber string) {
	if ProductSerialNumber == "" {
		activeCallFuncs.Range(func(_, value any) bool {
			value.(func())()
			return true
		})
		return
	}
	value, ok := activeCallFuncs.Load(ProductSerialNumber)
	if ok {
		value.(func())()
	}
}

func CheckLicense(l License) error {
	if l.AccessCode != accessCode {
		return errors.New("许可证与授权码不匹配")
	}
	if l.Endtime < time.Now().UnixMilli() {
		return errors.New("许可证到期")
	}
	return nil
}

func InstallReNewCheck() int64 {
	licenses := GetLicense()
	var installCount int64 = 0
	for _, v := range licenses {
		if v.AccessCode != accessCode {
			log.Print("许可证与授权码不匹配")
			continue
		}

		var isExp bool = false
		if v.Endtime < time.Now().UnixMilli() {
			isExp = true
		}
		// 通知
		if !isExp {
			installCount++
			notifyReNew(v.ProductNumber)
		}
	}
	return installCount
}

// 授权码计算逻辑
func GetAccessCode() string {
	accesscode := GetMac()
	hostname, _ := os.Hostname()
	accesscode = accesscode + hostname
	accesscode = system.MD5(accesscode)
	return accesscode
}

func GetAccessCode_V2() string {
	accesscode := GetMac()
	hostname, _ := os.Hostname()
	accesscode = accesscode + hostname
	accesscode = system.MD5(accesscode)
	return accesscode
}

// 注册超过有效期回调
func RegExpirationCall(ProductSerialNumber string, callbackfunc func()) {
	expirationCallFuncs.Store(ProductSerialNumber, callbackfunc)
}

// 产品续存回调
func RegReNewCall(ProductSerialNumber string, callbackfunc func()) {
	activeCallFuncs.Store(ProductSerialNumber, callbackfunc)
}

func GetMac() string {
	goos := runtime.GOOS
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v\n", err)
	}
	sort.Slice(netInterfaces, func(i, j int) bool {
		return netInterfaces[i].Index < netInterfaces[j].Index
	})
	switch goos {
	case "windows":
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}
			if strings.HasPrefix(strings.ToLower(netInterface.Name), "以太网") {
				return macAddr
			}
		}
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}

			if strings.HasPrefix(strings.ToLower(netInterface.Name), "eth") {
				return macAddr
			}
		}
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}
			if strings.HasPrefix(strings.ToLower(netInterface.Name), "wl") {
				return macAddr
			}
		}
	case "linux":
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}
			if strings.HasPrefix(strings.ToLower(netInterface.Name), "en") {
				return macAddr
			}
		}
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}
			if strings.HasPrefix(strings.ToLower(netInterface.Name), "et") {
				return macAddr
			}
		}
		for _, netInterface := range netInterfaces {
			macAddr := netInterface.HardwareAddr.String()
			if len(macAddr) == 0 {
				continue
			}
			if strings.HasPrefix(strings.ToLower(netInterface.Name), "wla") {
				return macAddr
			}
		}
	default:
		return ""
	}
	return ""
}

func GetLicense() []License {
	list, _ := ListDir(filepath.Join(basePath, "license"), "license")

	var listLicenses []License
	var id int64 = 1
	for _, v := range list {
		l, err := OpenLicense(v)

		if err != nil {
			log.Print(err.Error())
			continue
		}
		err = CheckLicense(l)
		if err != nil {
			l.ValidErrMessage = err.Error()
			l.Valid = false
		} else {
			l.Valid = true
		}
		l.ID = id
		id++
		listLicenses = append(listLicenses, l)
	}
	return listLicenses
}

func GetLicenseByProductNumber(productNumber string) []License {
	list, _ := ListDir(filepath.Join(basePath, "license"), "license")

	var listLicenses []License
	for _, v := range list {
		l, err := OpenLicense(v)

		if err != nil {
			log.Print(err.Error())
			continue
		}

		if l.ProductNumber != productNumber {
			continue
		}

		err = CheckLicense(l)
		if err != nil {
			l.ValidErrMessage = err.Error()
			l.Valid = false
		} else {
			l.Valid = true
		}
		listLicenses = append(listLicenses, l)
	}
	return listLicenses
}

func GetLicenseByName(name string) (License, error) {
	l, err := OpenLicense(filepath.Join(basePath, "license", name+".license"))
	if err != nil {
		return l, err
	}
	err = CheckLicense(l)
	if err != nil {
		l.ValidErrMessage = err.Error()
		l.Valid = false
	} else {
		l.Valid = true
	}
	return l, err
}

func QueryLicense(name string) (License, error) {
	list, _ := ListDir(filepath.Join(basePath, "license"), "license")
	for _, v := range list {
		_, fileName := filepath.Split(v)
		if name+".license" == fileName {
			return OpenLicense(v)
		}
	}
	var l License
	return l, errors.New("license not found")
}

func OpenLicense(path string) (License, error) {
	//只读打开
	var l License
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return l, err
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		return l, err
	}

	s, _ := system.DecryptPublic(b)

	err = json.Unmarshal([]byte(s), &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

func DecodeLicense(active_credential string) (License, error) {
	//只读打开
	var l License

	s, _ := system.DecryptPublic([]byte(active_credential))
	err := json.Unmarshal([]byte(s), &l)
	if err != nil {
		return l, err
	}
	return l, nil
}

// 自动激活
func AutoFristActive() {
	var conf struct {
		LicenseServerUrl string
		LicenseKey       string
		ProductNumber    string
	}
	system.ScanConfigJson(serverPath, &conf)
	if conf.LicenseKey == "" {
		return
	}
	licenses := GetLicense()
	for _, v := range licenses {
		if v.LicenseKey == conf.LicenseKey {
			if v.Valid {
				//许可证存在并且有效即不需要激活
				return
			}
		}
	}

	var active ActiveRequestInfo
	active.AccessCode = GetAccessCode()
	active.LicenseKey = conf.LicenseKey
	active.ProductNumber = conf.ProductNumber

	jb, _ := json.Marshal(active)

	b, err := system.EncryptPublic(jb)
	if err != nil {
		log.Print(err.Error())
		return
	}

	request := jsonrpc.NewJsonrpcrequest()
	request.Method = "/license/getActiveCredential"
	var activeRequest struct {
		ActiveCode string `json:"active_code"`
	}
	activeRequest.ActiveCode = string(b)
	request.Params = activeRequest

	rpcClient := jsonrpc.NewHttpClient()
	rpcClient.ErrorFunc = func(err error) {
		log.Print(err.Error())
	}
	rpcClient.Call(conf.LicenseServerUrl+"/jsonrpc", request, func(j jsonrpc.Jsonrpcresponse) {
		if j.Error.Code != 200 {
			log.Print(j.Error.Message)
			return
		}
		var result struct {
			ActiveCredential string `json:"active_credential"`
		}
		j.ReadResult(&result)
		err = SaveLicense(filepath.Join(basePath, "license", conf.ProductNumber+".license"), result.ActiveCredential)
		if err != nil {
			log.Print(j.Error.Message)
			return
		}
		InstallReNewCheck()
		_, err := GetLicenseByName(conf.ProductNumber)
		if err != nil {
			log.Print(j.Error.Message)
			return
		}
	})

}

func SaveLicense(savePath string, active string) error {
	file, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE, 0777) // 此处假设当前目录下已存在test目录
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(active)
	if err != nil {
		return err
	}
	return nil
}

// 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)

	dir, err := os.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

// 获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, _ error) error { //遍历目录

		if fi.IsDir() { // 忽略目录
			return nil
		}

		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}

		return nil
	})

	return files, err
}

var mbTable = []uint16{
	0x0000, 0x1189, 0x2312, 0x329b, 0x4624, 0x57ad, 0x6536, 0x74bf,
	0x8c48, 0x9dc1, 0xaf5a, 0xbed3, 0xca6c, 0xdbe5, 0xe97e, 0xf8f7,
	0x1081, 0x0108, 0x3393, 0x221a, 0x56a5, 0x472c, 0x75b7, 0x643e,
	0x9cc9, 0x8d40, 0xbfdb, 0xae52, 0xdaed, 0xcb64, 0xf9ff, 0xe876,
	0x2102, 0x308b, 0x0210, 0x1399, 0x6726, 0x76af, 0x4434, 0x55bd,
	0xad4a, 0xbcc3, 0x8e58, 0x9fd1, 0xeb6e, 0xfae7, 0xc87c, 0xd9f5,
	0x3183, 0x200a, 0x1291, 0x0318, 0x77a7, 0x662e, 0x54b5, 0x453c,
	0xbdcb, 0xac42, 0x9ed9, 0x8f50, 0xfbef, 0xea66, 0xd8fd, 0xc974,
	0x4204, 0x538d, 0x6116, 0x709f, 0x0420, 0x15a9, 0x2732, 0x36bb,
	0xce4c, 0xdfc5, 0xed5e, 0xfcd7, 0x8868, 0x99e1, 0xab7a, 0xbaf3,
	0x5285, 0x430c, 0x7197, 0x601e, 0x14a1, 0x0528, 0x37b3, 0x263a,
	0xdecd, 0xcf44, 0xfddf, 0xec56, 0x98e9, 0x8960, 0xbbfb, 0xaa72,
	0x6306, 0x728f, 0x4014, 0x519d, 0x2522, 0x34ab, 0x0630, 0x17b9,
	0xef4e, 0xfec7, 0xcc5c, 0xddd5, 0xa96a, 0xb8e3, 0x8a78, 0x9bf1,
	0x7387, 0x620e, 0x5095, 0x411c, 0x35a3, 0x242a, 0x16b1, 0x0738,
	0xffcf, 0xee46, 0xdcdd, 0xcd54, 0xb9eb, 0xa862, 0x9af9, 0x8b70,
	0x8408, 0x9581, 0xa71a, 0xb693, 0xc22c, 0xd3a5, 0xe13e, 0xf0b7,
	0x0840, 0x19c9, 0x2b52, 0x3adb, 0x4e64, 0x5fed, 0x6d76, 0x7cff,
	0x9489, 0x8500, 0xb79b, 0xa612, 0xd2ad, 0xc324, 0xf1bf, 0xe036,
	0x18c1, 0x0948, 0x3bd3, 0x2a5a, 0x5ee5, 0x4f6c, 0x7df7, 0x6c7e,
	0xa50a, 0xb483, 0x8618, 0x9791, 0xe32e, 0xf2a7, 0xc03c, 0xd1b5,
	0x2942, 0x38cb, 0x0a50, 0x1bd9, 0x6f66, 0x7eef, 0x4c74, 0x5dfd,
	0xb58b, 0xa402, 0x9699, 0x8710, 0xf3af, 0xe226, 0xd0bd, 0xc134,
	0x39c3, 0x284a, 0x1ad1, 0x0b58, 0x7fe7, 0x6e6e, 0x5cf5, 0x4d7c,
	0xc60c, 0xd785, 0xe51e, 0xf497, 0x8028, 0x91a1, 0xa33a, 0xb2b3,
	0x4a44, 0x5bcd, 0x6956, 0x78df, 0x0c60, 0x1de9, 0x2f72, 0x3efb,
	0xd68d, 0xc704, 0xf59f, 0xe416, 0x90a9, 0x8120, 0xb3bb, 0xa232,
	0x5ac5, 0x4b4c, 0x79d7, 0x685e, 0x1ce1, 0x0d68, 0x3ff3, 0x2e7a,
	0xe70e, 0xf687, 0xc41c, 0xd595, 0xa12a, 0xb0a3, 0x8238, 0x93b1,
	0x6b46, 0x7acf, 0x4854, 0x59dd, 0x2d62, 0x3ceb, 0x0e70, 0x1ff9,
	0xf78f, 0xe606, 0xd49d, 0xc514, 0xb1ab, 0xa022, 0x92b9, 0x8330,
	0x7bc7, 0x6a4e, 0x58d5, 0x495c, 0x3de3, 0x2c6a, 0x1ef1, 0x0f78,
}

func checkSum(data []byte) uint16 {
	var crc16 uint16
	crc16 = 0x0000
	for _, v := range data {
		n := uint8(uint16(v) ^ crc16)
		crc16 >>= 8
		crc16 ^= mbTable[n]
	}
	return crc16
}

func cehckSumString(data []byte) string {
	str := strconv.FormatInt(int64(checkSum(data)), 16)
	stLen := len(str)
	for i := 0; i < 4-stLen; i++ {
		str = "0" + str
	}
	return str
}

func SetProductNumber(p string) {
	productNumber = p
}

func getUUID() (string, error) {
	var cmd *exec.Cmd
	platform := strings.ToLower(osName())
	switch platform {
	case "windows":
		cmd = exec.Command("wmic", "csproduct", "get", "UUID")
	case "linux":
		cmd = exec.Command("cat", "/etc/machine-id")
	case "darwin":
		cmd = exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	uuid := parseUUID(string(out), platform)
	return uuid, nil
}

func osName() string {
	return fmt.Sprintf("%s", strings.ToLower(runtime.GOOS))
}

func parseUUID(input, platform string) string {
	switch platform {
	case "windows":
		lines := strings.Split(input, "\n")
		for _, line := range lines {
			if strings.Contains(line, "UUID") {
				return strings.TrimSpace(strings.Replace(line, "UUID", "", 1))
			}
		}
	case "linux":
		return strings.TrimSpace(input)
	case "darwin":
		lines := strings.Split(input, "\n")
		for _, line := range lines {
			if strings.Contains(line, "IOPlatformUUID") {
				return strings.TrimSpace(strings.Split(line, "=")[1])
			}
		}
	}
	return ""
}

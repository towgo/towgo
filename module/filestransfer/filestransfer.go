package filestransfer

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/system"
)

const (
	metaSuffix          = ".meta"
	dataSuffix          = ".data"
	queryStringKey      = "filekey"
	READ_JURISDICTION   = "READ_JURISDICTION"
	WRITE_JURISDICTION  = "WRITE_JURISDICTION"
	DELETE_JURISDICTION = "DELETE_JURISDICTION"
)

var allowUploadExt map[string]bool = map[string]bool{}

var contentMap map[string]string = map[string]string{
	".css":  "text/css",
	".html": "text/html",
	".js":   "application/javascript",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".json": "application/json",
	".gif":  "image/gif",
	".zip":  "application/x-zip-compressed",
	".rar":  "application/octet-stream",
	".txt":  "text/plain",
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/msword",
}

var crc8Table = []byte{
	0x00, 0x31, 0x62, 0x53, 0xc4, 0xf5, 0xa6, 0x97, 0xb9, 0x88, 0xdb, 0xea, 0x7d, 0x4c, 0x1f, 0x2e,
	0x43, 0x72, 0x21, 0x10, 0x87, 0xb6, 0xe5, 0xd4, 0xfa, 0xcb, 0x98, 0xa9, 0x3e, 0x0f, 0x5c, 0x6d,
	0x86, 0xb7, 0xe4, 0xd5, 0x42, 0x73, 0x20, 0x11, 0x3f, 0x0e, 0x5d, 0x6c, 0xfb, 0xca, 0x99, 0xa8,
	0xc5, 0xf4, 0xa7, 0x96, 0x01, 0x30, 0x63, 0x52, 0x7c, 0x4d, 0x1e, 0x2f, 0xb8, 0x89, 0xda, 0xeb,
	0x3d, 0x0c, 0x5f, 0x6e, 0xf9, 0xc8, 0x9b, 0xaa, 0x84, 0xb5, 0xe6, 0xd7, 0x40, 0x71, 0x22, 0x13,
	0x7e, 0x4f, 0x1c, 0x2d, 0xba, 0x8b, 0xd8, 0xe9, 0xc7, 0xf6, 0xa5, 0x94, 0x03, 0x32, 0x61, 0x50,
	0xbb, 0x8a, 0xd9, 0xe8, 0x7f, 0x4e, 0x1d, 0x2c, 0x02, 0x33, 0x60, 0x51, 0xc6, 0xf7, 0xa4, 0x95,
	0xf8, 0xc9, 0x9a, 0xab, 0x3c, 0x0d, 0x5e, 0x6f, 0x41, 0x70, 0x23, 0x12, 0x85, 0xb4, 0xe7, 0xd6,
	0x7a, 0x4b, 0x18, 0x29, 0xbe, 0x8f, 0xdc, 0xed, 0xc3, 0xf2, 0xa1, 0x90, 0x07, 0x36, 0x65, 0x54,
	0x39, 0x08, 0x5b, 0x6a, 0xfd, 0xcc, 0x9f, 0xae, 0x80, 0xb1, 0xe2, 0xd3, 0x44, 0x75, 0x26, 0x17,
	0xfc, 0xcd, 0x9e, 0xaf, 0x38, 0x09, 0x5a, 0x6b, 0x45, 0x74, 0x27, 0x16, 0x81, 0xb0, 0xe3, 0xd2,
	0xbf, 0x8e, 0xdd, 0xec, 0x7b, 0x4a, 0x19, 0x28, 0x06, 0x37, 0x64, 0x55, 0xc2, 0xf3, 0xa0, 0x91,
	0x47, 0x76, 0x25, 0x14, 0x83, 0xb2, 0xe1, 0xd0, 0xfe, 0xcf, 0x9c, 0xad, 0x3a, 0x0b, 0x58, 0x69,
	0x04, 0x35, 0x66, 0x57, 0xc0, 0xf1, 0xa2, 0x93, 0xbd, 0x8c, 0xdf, 0xee, 0x79, 0x48, 0x1b, 0x2a,
	0xc1, 0xf0, 0xa3, 0x92, 0x05, 0x34, 0x67, 0x56, 0x78, 0x49, 0x1a, 0x2b, 0xbc, 0x8d, 0xde, 0xef,
	0x82, 0xb3, 0xe0, 0xd1, 0x46, 0x77, 0x24, 0x15, 0x3b, 0x0a, 0x59, 0x68, 0xff, 0xce, 0x9d, 0xac,
}

var crc16Table = []uint16{
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

var dataBasePath string = "./fileobjects"
var crcBitType int = 8

type Account struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

//var isEncryption bool //是否加密

func SetDataBasePath(path string) {
	dataBasePath = path
}

// crc文件夹散列位数  支持8/16
func SetCrcBitType(bit int) {
	switch bit {
	case 16:
		crcBitType = bit
	default:
		crcBitType = 8
	}
}

type FilesObject struct {
	Files []File `json:"files"`
}

func (fo *FilesObject) AddByFileKey(fileKey string) error {
	file := File{}
	err := file.Load(fileKey)
	if err != nil {
		return err
	}
	fo.Files = append(fo.Files, file)
	return nil
}

func (fo *FilesObject) SaveAll() {
	for k, _ := range fo.Files {
		fo.Files[k].Save()
	}
}

func (f *File) getFolderCrcByFileKey(fileKey string) string {
	return f.GetCrcString([]byte(fileKey))
}

func (fo *File) crc16Sum(data []byte) uint16 {
	var crc16 uint16
	crc16 = 0x0000
	for _, v := range data {
		n := uint8(uint16(v) ^ crc16)
		crc16 >>= 8
		crc16 ^= crc16Table[n]
	}
	return crc16
}

func (fo *File) crc8Sum(buf []byte) (val uint8) {
	for i := len(buf) - 1; i >= 0; i-- {
		val = crc8Table[(val^buf[i])&0xff]
	}
	return
}

func (fo *File) GetCrcString(data []byte) string {
	var crc int64
	switch crcBitType {
	case 8:
		crc = int64(fo.crc8Sum(data))
	case 16:
		crc = int64(fo.crc16Sum(data))
	}
	return strconv.FormatInt(crc, 16)
}

func (FileKey) TableName() string {
	return "file_key"
}

type FileKey struct {
	FileKey     string    `json:"file_key" xorm:"pk" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Creator     string    `json:"creator"`
	Owners      []Account `json:"owners"`
	DwonloadUrl string    `json:"download_url"`
	UploadType  string    `json:"upload_type"`
	CreatedAt   int64     `json:"created_at" xorm:"created"`
	UpdatedAt   int64     `json:"updated_at" xorm:"updated"`
}

func (f *FileKey) AfterDelete(session basedboperat.DbTransactionSession) error {
	var file File
	file.FileKey = f.FileKey
	err := file.Delete()
	if err != nil {
		session.Rollback()
	}
	return err
}

type File struct {
	Name                   string   `json:"name"`
	Suffix                 string   `json:"suffix"`
	Data                   string   `json:"data" gob:"-"`
	FileKey                string   `json:"file_key"`
	EncodeType             string   `json:"encode_type"` //编码类型
	DwonloadUrl            string   `json:"download_url"`
	Creator                string   `json:"creator"`
	OwnerUsers             []int64  `json:"owner_users"`
	ReadJurisdictionUsers  []int64  `json:"read_jurisdiction_users"`
	WriteJurisdictionUsers []int64  `json:"write_jurisdiction_users"`
	CrcFolder              string   `json:"-"`
	inited                 bool     `json:"-" gob:"-"`
	fileHandller           *os.File `json:"-" gob:"-"`
}

func (f *File) GetDataStream() (stream []byte, err error) {
	if f.Data == "" {
		if f.FileKey == "" {
			return nil, errors.New("please set filekey first")
		}
		err := f.Load(f.FileKey)
		if err != nil {
			return nil, err
		}
	}
	stream = []byte(f.Data)
	return
}

// 写入前初始化
func (f *File) InitForSave(fileKey string) error {
	//初始化
	if fileKey != "" {
		f.FileKey = fileKey
		f.Delete()
	} else {
		f.FileKey = system.SHA1(system.GetGUID().Hex())
	}

	f.CrcFolder = f.getFolderCrcByFileKey(f.FileKey)
	f.GenerateUrl()
	f.createDateDir(f.CrcFolder)
	err := f.saveMetaObject()
	if err != nil {
		return err
	}
	f.inited = true
	return nil
}

func (f *File) DataSwitch() {
	dataStruct_a := strings.Split(f.Data, ":")
	if len(dataStruct_a) > 1 {
		dataStruct_b := strings.Split(dataStruct_a[1], ";")
		if len(dataStruct_b) > 1 {
			dataStruct_c := strings.Split(dataStruct_b[1], ",")
			if len(dataStruct_c) > 1 {
				f.EncodeType = dataStruct_c[0]
				f.Data = dataStruct_c[1]
			}
		}
	}
}

// 基于字符串编码的文件保存（base64等）
func (f *File) Save() error {
	f.InitForSave("")
	if f.Data == "" {
		return errors.New("数据不能为空")
	}
	f.DataSwitch()
	f.Write([]byte(f.Data))
	return nil
}

// 保存文件元信息文件
func (f *File) saveMetaObject() error {

	newFile := File{}
	newFile = *f
	newFile.Data = ""
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	enc.Encode(newFile)
	savePath := filepath.Join(dataBasePath, f.CrcFolder, f.FileKey)
	savePath = savePath + metaSuffix

	_, err := os.Stat(savePath)
	if os.IsExist(err) {
		os.Remove(savePath)
	}

	file, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777) // 此处假设当前目录下已存在test目录
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, &buf)
	if err != nil {
		return err
	}
	return nil
}

// 保存实体数据文件
func (f *File) saveDataObject(b []byte) (int, error) {
	savePath := filepath.Join(dataBasePath, f.CrcFolder, f.FileKey)
	savePath = savePath + dataSuffix

	_, err := os.Stat(savePath)
	if os.IsExist(err) {
		os.Remove(savePath)
	}

	file, err := os.OpenFile(savePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)

	if err != nil {
		return 0, err
	}

	defer file.Close()
	count, err := file.Write(b)

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (f *File) Load(fileKey string) error {
	f.CrcFolder = f.getFolderCrcByFileKey(fileKey)
	savePath := filepath.Join(dataBasePath, f.CrcFolder, fileKey)
	savePath = savePath + metaSuffix
	file, err := os.Open(savePath)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewDecoder(file)
	err = enc.Decode(f)
	if err != nil {
		return err
	}

	buf, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	switch f.EncodeType {
	case "binary": //二进制类型  进行base64转码
		f.Data = base64.StdEncoding.EncodeToString(buf)
	default: //base64
		f.Data = string(buf)
	}

	return nil
}

func (f *File) HasJurisdiction(jurisdictionType string, userid int64) bool {
	switch jurisdictionType {
	case READ_JURISDICTION:

	case WRITE_JURISDICTION:
	}
	return false
}

func (f *File) SetJurisdiction() {

}

// 生成文件下载链接
func (f *File) GenerateUrl() {
	f.DwonloadUrl = http_download_path + "?" + queryStringKey + "=" + f.FileKey
}

// 根据属主删除文件
func (f *File) DeleteByOwner(owner int64) error {
	findFile := File{}
	err := findFile.Load(f.FileKey)
	if err != nil {
		return err
	}
	if !f.HasJurisdiction(DELETE_JURISDICTION, owner) {
		return errors.New("权限不足,无法删除")
	}

	return findFile.delete()
}

func (f *File) Delete() error {
	findFile := File{}
	err := findFile.Load(f.FileKey)
	if err != nil {
		return err
	}
	return findFile.delete()
}

func (f *File) delete() error {

	savePath := filepath.Join(dataBasePath, f.CrcFolder, f.FileKey)
	os.Remove(savePath + metaSuffix)
	os.Remove(savePath + dataSuffix)
	return nil
}

func (f *File) createDateDir(path string) (folderPath string, folderName string) {
	folderPath = filepath.Join(dataBasePath, path)
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, os.ModePerm)
	}
	return
}

func (f *File) Read(b []byte) (int, error) {
	if f.fileHandller == nil {
		f.CrcFolder = f.getFolderCrcByFileKey(f.FileKey)
		savePath := filepath.Join(dataBasePath, f.CrcFolder, f.FileKey)
		savePath = savePath + dataSuffix
		file, err := os.Open(savePath)
		if err != nil {
			return 0, err
		}
		f.fileHandller = file
	}
	readCount, err := f.fileHandller.Read(b)
	if err != nil {
		f.fileHandller.Close()
		f.fileHandller = nil
	}
	return readCount, err
}

func (f *File) Close() error {
	if f.fileHandller == nil {
		return nil
	}
	return f.fileHandller.Close()
}

func (f *File) Write(b []byte) (int, error) {
	if !f.inited {
		return 0, errors.New("文件未初始化")
	}
	count, err := f.saveDataObject(b)
	return count, err
}

func (f *File) SaveDataTo(filePath string) error {
	if f.Data == "" {
		return errors.New("data is null")
	}

	b, err := base64.StdEncoding.DecodeString(f.Data)
	if err != nil {
		return err
	}

	_, err = os.Stat(filePath)
	if os.IsExist(err) {
		os.Remove(filePath)
	}

	fs, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer fs.Close()

	buf := bytes.Buffer{}
	buf.Write(b)
	io.Copy(fs, &buf)

	return nil
}

func AddAllowExt(exts []string) {
	for _, v := range exts {
		allowUploadExt[v] = true
	}
}

package tcfg

import (
	"bytes"
	"encoding/json"
	"github.com/towgo/towgo/errors/tcode"
	"github.com/towgo/towgo/errors/terror"
	"github.com/towgo/towgo/os/tfile"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultConfigFileName = "config.json" // 更改为更具描述性的默认文件名
	DefaultConfigFilePath = "config"      // 相对路径，更灵活
)

type Adapter struct {
	defaultFileNameOrPath string
	jsonData              map[string]interface{}
}

// NewAdapter 创建一个新的 Adapter 实例
func NewAdapter(fileNameOrPath ...string) (*Adapter, error) {
	var usedFileNameOrPath string

	if len(fileNameOrPath) > 0 {
		usedFileNameOrPath = fileNameOrPath[0]
	} else {
		// 使用相对路径拼接，默认配置文件路径
		usedFileNameOrPath = filepath.Join(DefaultConfigFilePath, DefaultConfigFileName)
	}

	adapter := &Adapter{
		defaultFileNameOrPath: usedFileNameOrPath,
	}

	if err := adapter.verify(); err != nil {
		return nil, err
	}

	return adapter, nil
}

// verify 验证配置文件是否存在且可读
func (a *Adapter) verify(path ...string) error {
	verifyPath := a.defaultFileNameOrPath
	if len(path) > 0 {
		verifyPath = path[0]
	}
	_, err := os.Stat(verifyPath)
	if os.IsNotExist(err) {
		return terror.NewCode(tcode.CodeBusinessValidationFailed, "Config file does not exist : "+verifyPath)
	}
	if os.IsPermission(err) {
		return terror.NewCode(tcode.CodeBusinessValidationFailed, "No permission to read configuration files :"+verifyPath)
	}
	if err != nil {
		return terror.NewCode(tcode.CodeBusinessValidationFailed, "An error occurred while verifying the configuration file: "+verifyPath)

	}

	return nil
}

func (a *Adapter) SetConfigPath(path string) error {
	err := a.verify(path)
	if err != nil {
		return err
	}
	a.defaultFileNameOrPath = path
	return nil
}
func (a *Adapter) GetConfigPath() string {
	return a.defaultFileNameOrPath
}

// LoadConfig 读取并解析配置文件到嵌套的map结构
func (a *Adapter) LoadConfig() error {
	var out map[string]interface{}
	filePath := a.GetConfigPath()
	dataType := tfile.ExtName(filePath)
	if dataType != "json" {
		return terror.NewCode(tcode.CodeBusinessValidationFailed, "invalid config file format: "+filePath)
	}
	content := tfile.GetContents(filePath)
	decoder := json.NewDecoder(bytes.NewReader([]byte(content)))
	if err := decoder.Decode(&out); err != nil {
		return terror.NewCodef(tcode.CodeBusinessValidationFailed, "json decoding failed for content: %s, error: %+v", content, err)

	}
	a.jsonData = out
	return nil
}

// Data retrieves and returns all configuration data in current resource as map.
// Note that this function may lead lots of memory usage if configuration data is too large,
// you can implement this function if necessary.
func (a *Adapter) Data() map[string]interface{} {
	return a.jsonData
}

// Get retrieves and returns value by specified `pattern` in current resource.
// Pattern like:
// "x.y.z" for map item.
// "x.0.y" for slice item.
func (a *Adapter) Get(pattern string) (interface{}, error) {
	// 处理特殊情况：返回全部数据
	if pattern == "" || pattern == "." {
		// 返回数据的深拷贝防止外部修改
		return a.deepCopy(a.jsonData)
	}
	var def interface{}
	// 分割路径
	parts := strings.Split(pattern, ".")
	var current interface{} = a.jsonData

	// 遍历路径节点
	for _, part := range parts {
		switch val := current.(type) {
		case map[string]interface{}:
			// 处理 map 类型
			if nextVal, exists := val[part]; exists {
				current = nextVal
			} else {
				// 键不存在时返回默认值
				return def, nil
			}
		case []interface{}:
			// 处理数组类型
			idx, err := strconv.Atoi(part)
			if err != nil {

				return def, terror.NewCodef(tcode.CodeBusinessValidationFailed, "invalid index format err = %+v", err)
			}
			if idx < 0 || idx >= len(val) {

				return def, terror.NewCodef(tcode.CodeBusinessValidationFailed, "index out of range: %d", idx)
			}
			current = val[idx]
		default:
			// 不支持的类型
			return def, terror.NewCodef(tcode.CodeBusinessValidationFailed, "unsupported data type: %T", val)
		}
	}

	return current, nil
}

// deepCopy 创建数据的深拷贝
func (a *Adapter) deepCopy(data interface{}) (interface{}, error) {
	// 使用 JSON 序列化和反序列化实现深拷贝
	jsonData, _ := json.Marshal(data)
	return json.Unmarshal(jsonData, &data), nil
}

// GetDataToStruct 将获取到的数据解析到结构体
func (a *Adapter) GetDataToStruct(key string, pointstr interface{}) error {
	data, err := a.Get(key)
	if err != nil {
		return err
	}
	jsonData, _ := json.Marshal(data)
	err = json.Unmarshal(jsonData, pointstr)
	if err != nil {
		return terror.NewCodef(tcode.CodeBusinessValidationFailed, "json Unmarshal err %+v", err)
	}
	return nil
}

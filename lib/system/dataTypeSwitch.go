package system

import (
	"errors"
	"reflect"
)

func ToInterfaceSlice(input interface{}) ([]interface{}, error) {
	// 获取输入值的反射值
	val := reflect.ValueOf(input)

	// 检查是否为切片类型
	if val.Kind() != reflect.Slice {
		return nil, errors.New("非切片类型无法转换") // 返回实际类型作为错误信息
	}

	// 创建结果切片
	result := make([]interface{}, val.Len())

	// 遍历原始切片并填充结果
	for i := 0; i < val.Len(); i++ {
		result[i] = val.Index(i).Interface()
	}

	return result, nil
}

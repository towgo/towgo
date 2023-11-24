/*
通用算法
*/
package system

import (
	"reflect"
)

// 泛型切片删除方法
func ScliceRemoveByValue[T any](sclice []T, value T) []T {
	result := make([]T, 0)
	if reflect.TypeOf(value).String() == "string" {
		for _, v := range sclice {
			if reflect.ValueOf(value).String() != reflect.ValueOf(v).String() {
				result = append(result, v)
			}
		}
	} else {
		for _, v := range sclice {
			if reflect.ValueOf(value) != reflect.ValueOf(v) {
				result = append(result, v)
			}
		}
	}

	return result
}

// 泛型 判断切片是否存在某个元素
func IsInSclice[T any](sclice []T, value T) bool {
	if reflect.TypeOf(value).String() == "string" {
		for _, v := range sclice {
			if reflect.ValueOf(v).String() == reflect.ValueOf(value).String() {
				return true
			}
		}
	} else {
		for _, v := range sclice {

			if reflect.ValueOf(v) == reflect.ValueOf(value) {
				return true
			}
		}
	}

	return false
}

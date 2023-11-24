package types

import (
	"reflect"

	"xorm.io/xorm/schemas"
)

func StrToType(name string) reflect.Type {
	switch name {
	// case "number":
	// 	return schemas.Int64Type
	// case "input":
	// 	return schemas.StringType
	// case "date", "datetime", "time":
	// 	return schemas.TimeType
	default:
		return schemas.StringType
	}

}

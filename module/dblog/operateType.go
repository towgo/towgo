package dblog

import (
	"github.com/towgo/towgo/dao/basedboperat"
	"github.com/towgo/towgo/lib/api"
	"log"
	"strings"
)

type OperateType struct {
	Id          int64
	Type        int64  `json:"type"`
	Method      string `json:"method"`
	ServiceName string `json:"serviceName"`
	ModuleName  string `json:"moduleName"`
	MethodName  string `json:"method_name"`
}

func (OperateType) TableName() string {
	return "operate_type"
}
func NewOperateType(o int64, method string, serverName string, methodName string) OperateType {
	methodList := strings.Split(method, "/")
	moduleName := methodList[1]
	return OperateType{
		Type:        o,
		Method:      method,
		ModuleName:  moduleName,
		ServiceName: serverName,
		MethodName:  methodName,
	}
}
func NewCrudOperateType(method string, serverName string, methodName string, CRUD_FLAG ...string) (os []OperateType) {
	methodList := strings.Split(method, "/")
	moduleName := methodList[1]
	if len(CRUD_FLAG) == 0 {
		os = []OperateType{
			OperateType{
				Type:        QUERY,
				Method:      method + "/list",
				ModuleName:  moduleName,
				ServiceName: serverName,
				MethodName:  methodName + "列表",
			},
			OperateType{
				Type:        QUERY,
				Method:      method + "/detail",
				ModuleName:  moduleName,
				ServiceName: serverName,
				MethodName:  methodName + "详情",
			},
			OperateType{
				Type:        UPDATE,
				Method:      method + "/update",
				ModuleName:  moduleName,
				ServiceName: serverName,
				MethodName:  methodName + "更新",
			},
			OperateType{
				Type:        CREATE,
				Method:      method + "/create",
				ModuleName:  moduleName,
				ServiceName: serverName,
				MethodName:  methodName + "创建",
			},
			OperateType{
				Type:        DELETE,
				Method:      method + "/delete",
				ModuleName:  moduleName,
				ServiceName: serverName,
				MethodName:  methodName + "删除",
			},
		}

	} else {
		for _, v := range CRUD_FLAG {
			switch v {
			case api.CRUD_FLAG_CREATE:
				os = append(os, OperateType{
					Type:        CREATE,
					Method:      method + "/create",
					ModuleName:  moduleName,
					ServiceName: serverName,
					MethodName:  methodName + "创建",
				})
			case api.CRUD_FLAG_DELETE:
				os = append(os, OperateType{
					Type:        DELETE,
					Method:      method + "/delete",
					ModuleName:  moduleName,
					ServiceName: serverName,
					MethodName:  methodName + "删除",
				})
			case api.CRUD_FLAG_DETAIL:
				os = append(os, OperateType{
					Type:        QUERY,
					Method:      method + "/detail",
					ModuleName:  moduleName,
					ServiceName: serverName,
					MethodName:  methodName + "详情",
				})
			case api.CRUD_FLAG_LIST:
				os = append(os, OperateType{
					Type:        QUERY,
					Method:      method + "/list",
					ModuleName:  moduleName,
					ServiceName: serverName,
					MethodName:  methodName + "列表",
				})
			case api.CRUD_FLAG_UPDATE:
				os = append(os, OperateType{
					Type:        UPDATE,
					Method:      method + "/update",
					ModuleName:  moduleName,
					ServiceName: serverName,
					MethodName:  methodName + "更新",
				})
			}
		}
	}
	return os
}
func BatchInsert(osArgs ...OperateType) error {
	go func() {
		for i, _ := range osArgs {
			var ost OperateType
			err := basedboperat.Get(&ost, nil, "method =? and service_name =? and module_name = ? ", osArgs[i].Method, osArgs[i].ServiceName, osArgs[i].ModuleName)
			if err != nil {
				log.Println("BatchInsert OperateType", err)
				continue
			}
			if ost.Id != 0 {
				if ost.MethodName != osArgs[i].MethodName {
					ost.MethodName = osArgs[i].MethodName
					basedboperat.Update(&ost, []string{"method_name"}, "id =?", ost.Id)
				}
				continue
			}
			_, err = basedboperat.Create(&osArgs[i])
			if err != nil {
				continue
			}
		}

	}()
	return nil
}

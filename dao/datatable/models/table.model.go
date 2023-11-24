package models

import (
	"reflect"
	"time"
	"xorm.io/xorm/schemas"
)

type TableModel struct {
	TableName  string                 `json:"name"`
	TableTitle string                 `json:"title"`
	TableData  map[string]interface{} `json:"tableData"`
}

// 解析数据表
func (d *TableModel) ParseTable() *StrutTable {
	strutTable := &StrutTable{
		TableName: d.TableName,
		Comment:   d.TableTitle,
	}
	var fields []reflect.StructField
	fields = append(fields, reflect.StructField{Name: "Id", Type: schemas.Int64Type, Tag: reflect.StructTag("xorm:pk autoincr")})
	//for _, f := range d.Fields {
	//	fields = append(fields, f.generateField())
	//}
	fields = append(fields, reflect.StructField{Name: "CreatedAt", Type: schemas.Int64Type, Tag: reflect.StructTag("")})
	fields = append(fields, reflect.StructField{Name: "UpdatedAt", Type: schemas.Int64Type, Tag: reflect.StructTag("")})
	fields = append(fields, reflect.StructField{Name: "CreatorId", Type: schemas.Int64Type, Tag: reflect.StructTag("")})
	fields = append(fields, reflect.StructField{Name: "UpdaterId", Type: schemas.Int64Type, Tag: reflect.StructTag("")})
	strutTable.Fields = fields
	return strutTable
}

// 结构体表模型
type StrutTable struct {
	TableName string
	Comment   string
	Fields    []reflect.StructField
}

func (t *TableModel) InitUpdate() map[string]interface{} {
	delete(t.TableData, "id")
	t.TableData["edit_time"] = time.Now().Format("2006-01-02 15:04:05")
	return t.TableData
}

func (t *TableModel) InitCreate() map[string]interface{} {
	t.TableData["edit_time"] = time.Now().Format("2006-01-02 15:04:05")
	t.TableData["create_time"] = time.Now().Format("2006-01-02 15:04:05")
	return t.TableData
}

type FieldModel struct {
	Name         string `json:"name"`
	Title        string `json:"title"`
	Type         string `json:"types"`
	Describe     string `json:"describe"`
	DefaultValue string `json:"default_value"`
}

type TableFieldModel struct {
	Table string     `json:"name"`
	Field FieldModel `json:"field"`
}

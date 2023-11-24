package engines

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/towgo/towgo/dao/datatable/ddl/dialects"
	dialects2 "github.com/towgo/towgo/dao/datatable/ddl/dialects"
	"github.com/towgo/towgo/dao/datatable/ddl/tags"
	"github.com/towgo/towgo/dao/datatable/ddl/tags/utils"
	"github.com/towgo/towgo/dao/datatable/models"
	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"xorm.io/xorm/caches"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"
)

type Xorm struct{}

func (g *Xorm) CreateTable(table *models.TableModel) (bool, error) {
	db := xormDriver.DbMaster()
	tableName := table.TableName
	if ok, _ := db.Engine.IsTableExist(tableName); ok {
		return false, errors.New(db.Engine.DriverName() + ":" + fmt.Sprintf("%s 数据表已经存在", tableName))
	}
	dialect, err := dialects2.OpenDialect(db.Engine.DriverName(), db.Engine.DataSourceName())
	if err != nil {
		log.Println(err)
		return false, err
	}
	cacherMgr := caches.NewManager()
	mapper := names.NewCacheMapper(new(names.SnakeMapper))
	tagParser := tags.NewParser("xorm", dialect, mapper, mapper, cacherMgr)
	refTable, _ := tagParser.Parse(table.ParseTable())

	if refTable.AutoIncrement != "" && dialect.Features().AutoincrMode == dialects2.SequenceAutoincrMode {
		sqlStr, err := dialect.CreateSequenceSQL(context.Background(), db.Engine.DB(), utils.SeqName(tableName))
		if err != nil {
			log.Println(err)
			return false, err
		}
		if _, err := db.Engine.Exec(sqlStr); err != nil {
			log.Println(err)
			return false, err
		}
	}

	sqlStr, _, err := dialect.CreateTableSQL(context.Background(), db.Engine.DB(), refTable, tableName)
	if err != nil {
		log.Println(err)
		return false, err
	}
	if _, err := db.Engine.Exec(sqlStr); err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

func (g *Xorm) AddField(table *models.TableFieldModel) (bool, error) {
	db := xormDriver.DbMaster()
	if dial, err := dialects.OpenDialect(db.Engine.DriverName(), db.Engine.DataSourceName()); err == nil {
		var typeValue string
		switch table.Field.Type {
		case "number":
			typeValue = schemas.BigInt
		case "string":
			typeValue = schemas.Text
		case "float":
			typeValue = schemas.Double
		default:
			typeValue = schemas.Text
		}
		col := schemas.NewColumn(table.Field.Name, table.Field.Name, schemas.SQLType{Name: typeValue}, 0, 0, true)
		addSql := dial.AddColumnSQL(table.Table, col)
		if _, err := db.Engine.Exec(addSql); err != nil {
			return false, err
		}
	} else {
		return false, err
	}
	return true, nil
}

func (g *Xorm) DeleteField(table *models.TableFieldModel) (bool, error) {
	db := xormDriver.DbMaster()
	if dial, err := dialects.OpenDialect(db.Engine.DriverName(), db.Engine.DataSourceName()); err == nil {
		dropSql := dial.DropColumnSQL(table.Table, table.Field.Name)
		if _, err := db.Engine.Exec(dropSql); err != nil {
			return false, err
		}
	} else {
		return false, err
	}
	return true, nil
}

func (g *Xorm) UpdateTable(table *models.TableModel) (bool, error) {
	// todo 修改字段
	return true, nil
}
func (g *Xorm) DropTable(table *models.TableModel) (bool, error) {
	db := xormDriver.DbMaster()
	tableName := table.TableName
	err := db.Engine.DropTables(tableName)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}
func (g *Xorm) Create(tableModel *models.TableModel) (int64, error) {
	db := xormDriver.DbMaster()
	session := db.Engine.NewSession()
	session.Table(tableModel.TableName)
	newId, err := session.Insert(tableModel.InitCreate())
	if err != nil {
		return 0, err
	}
	return newId, nil
}
func (g *Xorm) Update(tableModel *models.TableModel) error {
	db := xormDriver.DbMaster()
	session := db.Engine.NewSession()
	session.Table(tableModel.TableName)
	whereMap := map[string]interface{}{"id": tableModel.TableData["id"]}
	_, err := session.Where(whereMap).Update(tableModel.InitUpdate())
	if err != nil {
		return err
	}
	return nil
}

func (g *Xorm) Delete(tableModel *models.TableModel) error {
	db := xormDriver.DbMaster()
	session := db.Engine.NewSession()
	session.Table(tableModel.TableName)
	_, err := session.Where(tableModel.TableData).Delete()
	if err != nil {
		return err
	}
	return nil
}
func (g *Xorm) Query(tableQuery *models.TableQuery) (interface{}, error) {
	db := xormDriver.DbMaster()
	session := db.Engine.NewSession()
	session.Table(tableQuery.TableName)

	return map[string]interface{}{}, nil
}

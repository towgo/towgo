package basedboperat

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/towgo/towgo/dao/ormDriver/xormDriver"
	"xorm.io/builder"
	"xorm.io/xorm"
)

type Xorm struct {
	ctx context.Context
	sync.Mutex
	currentSelectFields []string
	session             *xorm.Session
}

func (orm *Xorm) WithValue(key, value any) {
	if orm.ctx == nil {
		orm.ctx = context.Background()
	}
	orm.ctx = context.WithValue(orm.ctx, key, value)
}

func (orm *Xorm) Value(key any) any {
	if orm.ctx == nil {
		return nil
	}
	return orm.ctx.Value(key)
}
func (orm *Xorm) HasValue(key, value any) bool {
	return orm.Value(key) == value
}
func (orm *Xorm) First(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, FirstValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)

	cacheKey := GenerateCacheKey(destModel, PrimaryKey, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	count := len(selectFields)
	if count > 0 {
		session = session.Select(strings.Join(selectFields, ","))
	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}
	ok, err := session.Asc(PrimaryKey).Get(destModel)
	setCache(destModel, cacheKey, 0)
	if !ok {
		return err
	}
	return nil
}

// 获取最后条记录
func (orm *Xorm) Last(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, LastValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)

	cacheKey := GenerateCacheKey(destModel, PrimaryKey, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	count := len(selectFields)
	if count > 0 {

		session = session.Select(strings.Join(selectFields, ","))
	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}
	ok, err := session.Desc(PrimaryKey).Get(destModel)
	setCache(destModel, cacheKey, 0)
	if !ok {
		return err
	}

	return nil
}

// 获取记录
func (orm *Xorm) Get(destModel interface{}, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, GetValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)

	cacheKey := GenerateCacheKey(destModel, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	count := len(selectFields)
	if count > 0 {
		session = session.Select(strings.Join(selectFields, ","))
	} else {
		session = session.Select("*")
	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}

	_, err := session.Get(destModel)
	setCache(destModel, cacheKey, 0)
	return err
}

// 更新记录
func (orm *Xorm) Update(model interface{}, fields any, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, UpdateValue)
	var selectFieldsTmp []string

	if fields != nil {
		tp := reflect.TypeOf(fields)
		switch tp.String() {
		case "map[string]interface {}":
			selectFileds := fields.(map[string]interface{})
			for k, _ := range selectFileds {
				if k != "id" {
					selectFieldsTmp = append(selectFieldsTmp, k)
				}
			}

		case "[]string":
			selectFileds := fields.([]string)
			selectFieldsTmp = selectFileds
		case "string":
			selectFieldsTmp = append(selectFieldsTmp, fields.(string))
		default:
			return errors.New("不支持的fields类型:" + tp.String())
		}

	}
	orm.currentSelectFields = selectFieldsTmp

	err := reflectMethodCall(model, InputCheck, orm)
	if err != nil {
		return err
	}

	err = reflectMethodCall(model, UpdateCheck, orm)
	if err != nil {
		return err
	}

	err = reflectMethodCall(model, BeforeSave, orm)
	if err != nil {
		return err
	}
	err = reflectMethodCall(model, BeforeUpdate, orm)
	if err != nil {
		return err
	}
	defer reflectMethodCall(model, AfterSave, orm)
	defer reflectMethodCall(model, AfterUpdate, orm)
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbMaster()
		session = db.Engine.NewSession()
	}

	if fields != nil {

		tp := reflect.TypeOf(fields)
		switch tp.String() {
		case "map[string]interface {}":
			selectFileds := fields.(map[string]interface{})
			s := []string{}
			for k, _ := range selectFileds {
				if k != "id" {
					s = append(s, k)
				}
			}
			session = session.Cols(s...)
		case "[]string":
			selectFileds := fields.([]string)
			session = session.Cols(selectFileds...)
		case "string":
			session = session.Cols(fields.(string))
		default:
			return errors.New("不支持的fields类型:" + tp.String())
		}

	}
	if condition != nil {
		_, err = session.Where(condition, conditionArgs...).Update(model)
	} else {
		_, err = session.ID(ReflectStructID(model)).Update(model)
	}

	return err
}

// 删除记录
func (orm *Xorm) Delete(model interface{}, PrimaryKeyID interface{}, condition interface{}, conditionArgs ...interface{}) (int64, error) {
	orm.WithValue(DbOperateBeforeKey, DeleteValue)

	err := reflectMethodCall(model, BeforeDelete, orm)
	if err != nil {
		return 0, err
	}
	defer reflectMethodCall(model, AfterDelete, orm)
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbMaster()
		session = db.Engine.NewSession()
	}

	if PrimaryKeyID != nil {
		return session.ID(PrimaryKeyID).Delete(model)
	}
	if condition != nil {
		return session.Where(condition, conditionArgs...).Delete(model)
	}
	return session.ID(ReflectStructID(model)).Delete(model)
}

// 创建记录
func (orm *Xorm) Create(model interface{}) (int64, error) {
	orm.WithValue(DbOperateBeforeKey, CreateValue)

	err := reflectMethodCall(model, InputCheck, orm)
	if err != nil {
		return 0, err
	}
	err = reflectMethodCall(model, CreateCheck, orm)
	if err != nil {
		return 0, err
	}

	err = reflectMethodCall(model, BeforeCreate, orm)
	if err != nil {
		return 0, err
	}

	defer func() {
		reflectMethodCall(model, AfterCreate, orm)
	}()
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbMaster()
		session = db.Engine.NewSession()
	}

	return session.Insert(model)
}

// 执行原生sql语句
func (orm *Xorm) SqlExec(sql interface{}, args ...interface{}) error {
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbMaster()
		session = db.Engine.NewSession()
	}

	sqlsclie := []interface{}{sql}
	sqlsclie = append(sqlsclie, args...)
	_, err := session.Exec(sqlsclie...)
	return err
}

// 原生sql查询
func (orm *Xorm) SqlQuery(sql interface{}, args ...interface{}) (resultsSlice []map[string]interface{}, err error) {
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}
	sqlsclie := []interface{}{sql}
	sqlsclie = append(sqlsclie, args...)
	return session.QueryInterface(sqlsclie...)
}

/*
// 原生sql查询解析到结构体或MAP

	func (orm *Xorm) SqlQueryScan(destModel interface{}, sql interface{}, args ...interface{}) error {
		cacheKey := GenerateCacheKey(destModel, sql, args)
		if queryCache(destModel, cacheKey) {
			return nil
		}

		var session *xorm.Session
		if orm.session != nil {
			session = orm.session
		} else {
			db := xormDriver.DbSlave()
			session = db.Engine.NewSession()
		}

		sqlsclie := []interface{}{sql}
		sqlsclie = append(sqlsclie, args...)
		resultsSlice, err := session.QueryInterface(sqlsclie...)
		if err != nil {
			return err
		}

		b, err := json.Marshal(resultsSlice)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, destModel)
		//setCache(destModel, cacheKey, 0)
		return err
	}
*/
var (
	fieldNameCache = make(map[reflect.Type]map[string]string)
	cacheMutex     = sync.RWMutex{}
)

// SqlQueryScan 优化后的SQL查询解析函数，使用标准库实现命名转换
func (orm *Xorm) SqlQueryScan(destModel interface{}, sql interface{}, args ...interface{}) error {
	// 检查缓存
	cacheKey := GenerateCacheKey(destModel, sql, args)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	// 获取数据库会话
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	// 执行查询
	sqlSlice := []interface{}{sql}
	sqlSlice = append(sqlSlice, args...)
	resultsSlice, err := session.QueryInterface(sqlSlice...)
	if err != nil {
		return err
	}

	// 获取目标类型信息
	destValue := reflect.ValueOf(destModel)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destModel must be a pointer")
	}

	destValue = destValue.Elem()
	destType := destValue.Type()

	// 处理查询结果
	if len(resultsSlice) == 0 {
		return nil
	}

	// 如果是map类型，直接填充
	if destValue.Kind() == reflect.Map {
		return fillMap(destValue, resultsSlice[0])
	}

	// 处理结构体
	return fillStruct(destValue, destType, resultsSlice[0])
}

// fillMap 填充map类型数据
func fillMap(destValue reflect.Value, data map[string]interface{}) error {
	if destValue.IsNil() {
		destValue.Set(reflect.MakeMap(destValue.Type()))
	}

	for k, v := range data {
		destValue.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return nil
}

// fillStruct 填充结构体数据
func fillStruct(destValue reflect.Value, destType reflect.Type, data map[string]interface{}) error {
	// 获取字段名映射缓存
	fieldMap := getFieldMap(destType)

	for dbName, value := range data {
		// 查找对应的结构体字段
		fieldName, ok := fieldMap[dbName]
		if !ok {
			continue
		}

		// 设置字段值
		field := destValue.FieldByName(fieldName)
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// 类型转换
		val := reflect.ValueOf(value)
		if val.Type().ConvertibleTo(field.Type()) {
			field.Set(val.Convert(field.Type()))
		}
	}
	return nil
}

// getFieldMap 获取结构体字段名映射(缓存)
func getFieldMap(t reflect.Type) map[string]string {
	cacheMutex.RLock()
	if m, ok := fieldNameCache[t]; ok {
		cacheMutex.RUnlock()
		return m
	}
	cacheMutex.RUnlock()

	fieldMap := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 获取json标签或使用字段名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		} else {
			// 处理json标签中的选项如omitempty
			if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
				jsonTag = jsonTag[:commaIdx]
			}
		}

		// 将驼峰转为蛇形作为数据库字段名
		dbName := camelToSnake(jsonTag)
		fieldMap[dbName] = field.Name
	}

	cacheMutex.Lock()
	fieldNameCache[t] = fieldMap
	cacheMutex.Unlock()

	return fieldMap
}

// camelToSnake 将驼峰命名转为蛇形命名
func camelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// snakeToCamel 将蛇形命名转为驼峰命名
func snakeToCamel(s string) string {
	var result []rune
	upperNext := false
	for _, r := range s {
		if r == '_' {
			upperNext = true
		} else {
			if upperNext {
				result = append(result, unicode.ToUpper(r))
				upperNext = false
			} else {
				result = append(result, r)
			}
		}
	}
	return string(result)
}

// 执行根据条件查询
func (orm *Xorm) QueryScan(destModel interface{}, extra interface{}, condition interface{}, args ...interface{}) error {
	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	if condition != nil {
		session = session.Where(condition, args...)
	}
	if extra != nil {
		switch extra.(type) {
		case string:
			session = session.OrderBy(extra)
		case *ListSimple:
			l := extra.(*ListSimple)
			if len(l.Field) > 0 {
				session = session.Select(strings.Join(l.Field, ","))
			} else {
				session = session.Select("*")
			}
			if len(l.In) > 0 {
				for k, v := range l.In {
					if len(v) > 0 {
						session = session.In(k, v)
					}
				}
			}
			if l.Table != "" {
				session = session.Table(l.Table)
			}
			//排序
			for _, order := range l.Orderby {
				for k, v := range order {
					session = session.OrderBy(k + " " + v)
				}
			}
		}
	}

	err := session.Find(destModel)
	if err != nil {
		return err
	}
	return nil
}

func (orm *Xorm) ListScan(l *List, model interface{}, destModels interface{}) {
	orm.WithValue(DbOperateBeforeKey, ListScanValue)
	cacheKey := GenerateCacheKey(destModels, l, model)
	if queryCache(destModels, cacheKey) {
		queryCache(l, cacheKey+"list")

		reflectSliceModelCall(destModels, AfterQuery, orm)
		return
	}

	var session *xorm.Session
	var sessionCount *xorm.Session
	if orm.session != nil {
		session = orm.session
		sessionCount = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
		sessionCount = db.Engine.NewSession()
	}

	//统计总数
	var count int64

	//分页判断
	if l.Page <= 0 {
		l.Page = 1
	}
	if l.Limit == 0 {
		l.Limit = 10
	}
	offset := (l.Page - 1) * l.Limit
	//使用会话链
	dbSessionLink := session
	dbSessionLinkCount := sessionCount

	if len(l.Field) > 0 {
		dbSessionLink = dbSessionLink.Select(strings.Join(l.Field, ","))
	} else {
		dbSessionLink = dbSessionLink.Select("*")
	}

	if len(l.And) > 0 {
		for k, v := range l.And {
			if v == nil {
				//par := k + " is NULL"

				//dbSessionLinkCount = dbSessionLinkCount.Where(par)
				//dbSessionLink = dbSessionLink.Where(par)
				dbSessionLinkCount = dbSessionLinkCount.In(k, v...)
				dbSessionLink = dbSessionLink.In(k, v...)
			} else if len(v) > 0 {
				//par := k + " IN ?"
				//dbSessionLinkCount = dbSessionLinkCount.Where(par, v)
				//dbSessionLink = dbSessionLink.Where(par, v)
				dbSessionLinkCount = dbSessionLinkCount.In(k, v)
				dbSessionLink = dbSessionLink.In(k, v)
			}

		}
	}

	if len(l.Or) > 0 {
		for k, v := range l.Or {
			if v == nil {
				par := k + " is NULL"
				dbSessionLinkCount = dbSessionLinkCount.Or(par)
				dbSessionLink = dbSessionLink.Or(par)
			} else if len(v) > 0 {
				for _, vv := range v {
					par := k + " = ?"
					dbSessionLinkCount = dbSessionLinkCount.Or(par, vv)
					dbSessionLink = dbSessionLink.Or(par, vv)
				}
			}
		}
	}

	if len(l.Not) > 0 {
		for k, v := range l.Not {
			if v == nil {
				par := k + " is not NULL"
				dbSessionLinkCount = dbSessionLinkCount.Where(par)
				dbSessionLink = dbSessionLink.Where(par)
			} else if len(v) > 0 {
				dbSessionLinkCount = dbSessionLinkCount.NotIn(k, v)
				dbSessionLink = dbSessionLink.NotIn(k, v)
			}
		}
	}

	if len(l.Like) > 0 {
		for k, v := range l.Like {
			if len(v) > 0 {
				par := ""
				for i := 0; i < len(v); i++ {
					if par == "" {
						par = par + k + " LIKE ?"
					} else {
						par = par + " OR " + k + " LIKE ?"
					}
				}
				dbSessionLinkCount = dbSessionLinkCount.Where(par, v...)
				dbSessionLink = dbSessionLink.Where(par, v...)
			}
		}
	}

	if len(l.Where) > 0 {
		for _, v := range l.Where {
			// find_in_set 支持多值查询
			op := strings.ToLower(v.Operator)
			if op == "find_in_set" {
				cond := getFindInSet(v)
				if cond != nil {
					dbSessionLink = dbSessionLink.Where(cond)
					dbSessionLinkCount = dbSessionLinkCount.Where(cond)
				}
			} else {
				dbSessionLink = dbSessionLink.Where(v.Field+" "+v.Operator+" ?", v.Value)
				dbSessionLinkCount = dbSessionLinkCount.Where(v.Field+" "+v.Operator+" ?", v.Value)
			}
		}
	}

	//分页查询
	if l.Limit > 0 {
		dbSessionLink = dbSessionLink.Limit(l.Limit, offset)
	}

	//排序
	for _, order := range l.Orderby {
		for k, v := range order {
			dbSessionLink = dbSessionLink.OrderBy(k + " " + v)
		}
	}
	//dbSessionLink = dbSessionLink.Order("id desc")
	dbSessionLink = dbSessionLink.Table(model)
	//执行sql语句
	err := dbSessionLink.Find(destModels)
	if err != nil {
		log.Print(err.Error())
	}
	l.Error = err

	count, err = dbSessionLinkCount.Count(model)
	if err != nil {
		log.Print(err.Error())
	}

	l.Count = count

	//exp := getModelExpire(destModels)
	//setCache(destModels, cacheKey, exp)
	//setCache(l, cacheKey+"list", exp)
	reflectSliceModelCall(destModels, AfterQuery, orm)

}

func getFindInSet(v Condition) builder.Cond {
	var setValue []string
	switch v.Value.(type) {
	case string:
		setValue = strings.Split(v.Value.(string), ",")
	case float64:
		setValue = append(setValue, strconv.FormatFloat(v.Value.(float64), 'f', -1, 32))
	case []interface{}:
		for _, f := range v.Value.([]interface{}) {
			if value, ok := f.(string); ok {
				setValue = append(setValue, value)
			} else if value, ok := f.(float64); ok {
				setValue = append(setValue, strconv.FormatFloat(value, 'f', -1, 32))
			}
		}
	}
	if len(setValue) > 0 {
		var cons []builder.Cond
		for _, value := range setValue {
			cons = append(cons, builder.Expr("find_in_set(?,"+v.Field+")", value))
		}
		return builder.Or(cons...)
	}
	return nil
}

// 获取记录
func (orm *Xorm) Count(model interface{}, intPrt *int64, condition interface{}, conditionArgs ...interface{}) error {

	var session *xorm.Session
	if orm.session != nil {
		session = orm.session
	} else {
		db := xormDriver.DbSlave()
		session = db.Engine.NewSession()
	}

	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}

	count, err := session.Count(model)
	*intPrt = count
	return err
}

func (orm *Xorm) GetCurrentSelectFields() []string {
	return orm.currentSelectFields
}

func (orm *Xorm) IsCurrentSelectedField(field string) bool {
	for _, v := range orm.currentSelectFields {
		if v == field {
			return true
		}
	}
	return false
}

func (orm *Xorm) Sync(beans ...any) {
	xormDriver.Sync2(beans...)
}

func (orm *Xorm) Begin() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session != nil {
		return errors.New("xorm 事务已经开始 无法再次启动")
	}
	db := xormDriver.DbMaster()
	orm.session = db.Engine.NewSession()
	return orm.session.Begin()
}

func (orm *Xorm) Commit() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session == nil {
		return errors.New("xorm 事务未启动 无法提交")
	}
	err := orm.session.Commit()
	orm.session.Close()
	return err
}

func (orm *Xorm) Rollback() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session == nil {
		return errors.New("xorm 事务未启动 无法回滚")
	}
	err := orm.session.Rollback()
	orm.session.Close()

	orm.session = nil
	return err
}

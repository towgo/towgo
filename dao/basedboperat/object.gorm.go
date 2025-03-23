package basedboperat

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/towgo/towgo/dao/ormDriver/gormDriver"

	"gorm.io/gorm"
)

type Gorm struct {
	ctx context.Context
	sync.Mutex
	currentSelectFields []string
	session             *gorm.DB
}

func (orm *Gorm) WithValue(key, value any) {
	if orm.ctx == nil {
		orm.ctx = context.Background()
	}
	orm.ctx = context.WithValue(orm.ctx, key, value)
}

func (orm *Gorm) Value(key any) any {
	if orm.ctx == nil {
		return nil
	}
	return orm.ctx.Value(key)
}
func (orm *Gorm) HasValue(key, value any) bool {
	return orm.Value(key) == value
}
func (orm *Gorm) First(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, FirstValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)
	cacheKey := GenerateCacheKey(destModel, PrimaryKey, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}

	count := len(selectFields)
	if count > 0 {
		session = session.Select(selectFields)
	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}

	err := session.Order(PrimaryKey + " ASC").Last(destModel).Error
	setCache(destModel, cacheKey, 0)
	return err
}

// 获取最后条记录
func (orm *Gorm) Last(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, LastValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)

	cacheKey := GenerateCacheKey(destModel, PrimaryKey, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}

	count := len(selectFields)
	if count > 0 {
		session = session.Select(selectFields)
	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}

	err := session.Order(PrimaryKey + " DESC").Last(destModel).Error
	setCache(destModel, cacheKey, 0)
	return err
}

// 获取记录
func (orm *Gorm) Get(destModel interface{}, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, GetValue)
	orm.currentSelectFields = selectFields
	defer reflectMethodCall(destModel, AfterQuery, orm)

	cacheKey := GenerateCacheKey(destModel, selectFields, condition, conditionArgs)
	if queryCache(destModel, cacheKey) {
		return nil
	}

	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}

	count := len(selectFields)
	if count > 0 {
		session = session.Select(selectFields)

	}
	if condition != nil {
		session = session.Where(condition, conditionArgs...)
	}

	session = session.Find(destModel)
	setCache(destModel, cacheKey, 0)
	return session.Error
}

// 更新记录
func (orm *Gorm) Update(model interface{}, fields any, condition interface{}, conditionArgs ...interface{}) error {
	orm.WithValue(DbOperateBeforeKey, UpdateValue)
	var selectFieldsTmp []string

	if fields != nil {
		tp := reflect.TypeOf(fields)
		switch tp.String() {
		case "map[string]interface {}":
			selectFileds := fields.(map[string]interface{})
			for k, _ := range selectFileds {
				selectFieldsTmp = append(selectFieldsTmp, k)
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

	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbMaster()
		session = db.Engine
	}

	if fields != nil {

		tp := reflect.TypeOf(fields)
		switch tp.String() {
		case "map[string]interface {}":
			selectFileds := fields.(map[string]interface{})
			s := []string{}
			for k, _ := range selectFileds {
				s = append(s, k)
			}
			session = session.Select(s)
		case "[]string":
			selectFileds := fields.([]string)
			session = session.Select(selectFileds)
		case "string":
			session = session.Select(fields.(string))
		default:
			return errors.New("不支持的fields类型:" + tp.String())
		}

	}

	return session.Where(condition, conditionArgs...).Updates(model).Error
}

// 删除记录
func (orm *Gorm) Delete(model interface{}, PrimaryKeyID interface{}, condition interface{}, conditionArgs ...interface{}) (int64, error) {
	orm.WithValue(DbOperateBeforeKey, DeleteValue)

	err := reflectMethodCall(model, BeforeDelete, orm)
	if err != nil {
		return 0, err
	}
	defer reflectMethodCall(model, AfterDelete, orm)
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbMaster()
		session = db.Engine
	}

	if PrimaryKeyID != nil {
		session = session.Delete(model, PrimaryKeyID)
		return session.RowsAffected, session.Error
	}

	if condition != nil {
		session = session.Where(condition, conditionArgs...).Delete(model)
		return session.RowsAffected, session.Error
	}

	session = session.Delete(model)

	return session.RowsAffected, session.Error
}

// 创建记录
func (orm *Gorm) Create(model interface{}) (int64, error) {
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

	err = reflectMethodCall(model, BeforeSave, orm)
	if err != nil {
		return 0, err
	}
	defer func() {
		reflectMethodCall(model, AfterCreate, orm)
		reflectMethodCall(model, AfterSave, orm)
	}()
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbMaster()
		session = db.Engine
	}
	session = session.Create(model)
	return session.RowsAffected, session.Error
}

// 执行原生sql语句
func (orm *Gorm) SqlExec(sql interface{}, args ...interface{}) error {
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbMaster()
		session = db.Engine
	}
	return session.Exec(sql.(string), args...).Error
}

// 原生sql查询
func (orm *Gorm) SqlQuery(sql interface{}, args ...interface{}) (resultsSlice []map[string]interface{}, err error) {
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbMaster()
		session = db.Engine
	}
	session.Raw(sql.(string), args...).Scan(&resultsSlice)
	return
}

// 原生sql查询解析到结构体或MAP
func (orm *Gorm) SqlQueryScan(destModel interface{}, sql interface{}, args ...interface{}) error {
	cacheKey := GenerateCacheKey(destModel, sql, args)
	if queryCache(destModel, cacheKey) {
		return nil
	}
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}
	err := session.Raw(sql.(string), args...).Scan(destModel).Error
	//setCache(destModel, cacheKey, 0)
	return err
}

// 执行根据条件查询
func (orm *Gorm) QueryScan(destModel interface{}, extra interface{}, condition interface{}, args ...interface{}) error {
	//判断是否使用事务
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}
	if extra != nil {
		switch extra.(type) {
		case string:
			session = session.Order(extra)
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
						session = session.Where(k, v)
					}
				}
			}
			if l.Table != "" {
				session = session.Table(l.Table)
			}
			//排序
			for _, order := range l.Orderby {
				for k, v := range order {
					session = session.Order(k + " " + v)
				}
			}
		}
	}
	session.Where(condition, args...).Find(destModel)
	return nil
}

func (orm *Gorm) ListScan(l *List, model interface{}, destModels interface{}) {
	orm.WithValue(DbOperateBeforeKey, ListScanValue)
	cacheKey := GenerateCacheKey(destModels, l, model)

	if queryCache(destModels, cacheKey) {
		queryCache(l, cacheKey+"list")
		reflectSliceModelCall(destModels, AfterQuery, orm)
		return
	}

	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}

	//统计总数
	var count int64
	dbcount := session.Model(model)

	//分页判断
	if l.Page <= 0 {
		l.Page = 1
	}
	if l.Limit == 0 {
		l.Limit = 10
	}
	offset := (l.Page - 1) * l.Limit

	//使用会话链
	dbSessionLink := session.Model(model)

	//设定需要查询的字段
	if len(l.Field) > 0 {
		dbSessionLink = dbSessionLink.Select(l.Field)
	} else {
		dbSessionLink = dbSessionLink.Select("*")
	}

	if len(l.And) > 0 {
		for k, v := range l.And {
			if v == nil {
				par := k + " is NULL"
				dbcount = dbcount.Where(par)
				dbSessionLink = dbSessionLink.Where(par)
			} else if len(v) > 0 {
				par := k + " IN ?"
				dbcount = dbcount.Where(par, v)
				dbSessionLink = dbSessionLink.Where(par, v)
			}
		}
	}

	if len(l.Or) > 0 {
		for k, v := range l.Or {
			if v == nil {
				par := k + " is NULL"
				dbcount = dbcount.Or(par)
				dbSessionLink = dbSessionLink.Or(par)
			} else if len(v) > 0 {
				par := k + " IN ?"
				dbcount = dbcount.Or(par, v)
				dbSessionLink = dbSessionLink.Or(par, v)
			}
		}
	}

	if len(l.Not) > 0 {
		for k, v := range l.Not {
			if v == nil {
				par := k + " is not NULL"
				dbcount = dbcount.Where(par)
				dbSessionLink = dbSessionLink.Where(par)
			} else if len(v) > 0 {
				par := k + " IN ?"
				dbcount = dbcount.Not(par, v)
				dbSessionLink = dbSessionLink.Not(par, v)
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
				dbSessionLink = dbSessionLink.Where(par, v...)
				dbcount = dbcount.Where(par, v...)
			}
		}
	}

	if len(l.Where) > 0 {
		for _, v := range l.Where {
			dbSessionLink = dbSessionLink.Where(v.Field+" "+v.Operator+" ?", v.Value)
			dbcount = dbcount.Where(v.Field+" "+v.Operator+" ?", v.Value)
		}
	}

	//分页查询
	if l.Limit > 0 {
		dbSessionLink = dbSessionLink.Limit(l.Limit).Offset(offset)
	}

	//排序
	for _, order := range l.Orderby {
		for k, v := range order {
			dbSessionLink = dbSessionLink.Order(k + " " + v)
		}
	}

	//执行sql语句
	dbSessionLink = dbSessionLink.Find(destModels)
	l.Error = dbSessionLink.Error
	//查询总数
	dbcount.Count(&count)
	l.Count = count
	reflectSliceModelCall(destModels, AfterQuery, orm)

	if dbSessionLink.Error != nil {
		l.Error = dbSessionLink.Error
	}

}

func (orm *Gorm) Count(model interface{}, intPtr *int64, condition interface{}, args ...interface{}) error {
	var session *gorm.DB
	if orm.session != nil {
		session = orm.session
	} else {
		db := gormDriver.DbSlave()
		session = db.Engine
	}

	//统计总数
	var count int64
	session = session.Model(model)
	if condition != nil {
		session = session.Where(condition, args...)
	}
	session = session.Count(&count)
	if session.Error != nil {
		return session.Error
	}
	*intPtr = count
	return nil
}

func (orm *Gorm) GetCurrentSelectFields() []string {
	return orm.currentSelectFields
}

func (orm *Gorm) IsCurrentSelectedField(field string) bool {
	for _, v := range orm.currentSelectFields {
		if v == field {
			return true
		}
	}
	return false
}

func (orm *Gorm) Sync(beans ...any) {
	gormDriver.Sync2(beans...)
}

func (orm *Gorm) Begin() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session != nil {
		return errors.New("gorm 事务已经开始 无法再次启动")
	}
	orm.session = gormDriver.DbMaster().Engine.Begin()
	return nil
}

func (orm *Gorm) Commit() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session == nil {
		return errors.New("gorm 事务未启动 无法提交")
	}
	err := orm.session.Commit().Error
	orm.session = nil
	return err
}

func (orm *Gorm) Rollback() error {
	orm.Lock()
	defer orm.Unlock()
	if orm.session == nil {
		return errors.New("gorm 事务未启动 无法回滚")
	}
	err := orm.session.Rollback().Error
	orm.session = nil
	return err
}

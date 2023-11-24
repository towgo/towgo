/*
通用基本数据库操作
*/
package basedboperat

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/towgo/towgo/lib/system"
)

var globalCacheExpire int64 = 0
var queryCacheMap sync.Map

type packageStruc struct {
	mode   string
	engine DbOperat
	lock   sync.Mutex
}

func (p *packageStruc) Lock() {
	p.lock.Lock()
}

func (p *packageStruc) Unlock() {
	p.lock.Unlock()
}

func (p *packageStruc) SetOrmEngine(engineName string) error {
	p.Lock()
	switch engineName {
	case "xorm":
		p.engine = &Xorm{}
	case "gorm":
		p.engine = &Gorm{}
	default:
		return errors.New("不支持的引擎:" + engineName)
	}

	p.mode = engineName

	p.Unlock()
	return nil
}

func (p *packageStruc) GetEngine(ctx context.Context) DbOperat {
	switch p.mode {
	case "xorm":
		return &Xorm{ctx: ctx}
	case "gorm":
		return &Gorm{ctx: ctx}
	}
	return p.engine
}

func (p *packageStruc) NewEngineSession(ctx context.Context) (DbTransactionSession, error) {
	switch p.mode {
	case "xorm":
		return &Xorm{ctx: ctx}, nil
	case "gorm":
		return &Gorm{ctx: ctx}, nil
	default:
		return nil, errors.New("当前引擎不支持事务")
	}
}

var packageObject *packageStruc

func init() {
	packageObject = &packageStruc{
		mode:   "xorm",
		engine: &Xorm{},
	}
}

func SetGlobalCacheExpire(expire int64) {
	globalCacheExpire = expire
}

// 设定orm引擎  gorm|xorm|...
func SetOrmEngine(engineName string) error {
	return packageObject.SetOrmEngine(engineName)
}

// 获取第一条记录
func First(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.First(destModel, PrimaryKey, selectFields, condition, conditionArgs...)

}

// 获取最后条记录
func Last(destModel interface{}, PrimaryKey string, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Last(destModel, PrimaryKey, selectFields, condition, conditionArgs...)
}

// 获取记录
func Get(destModel interface{}, selectFields []string, condition interface{}, conditionArgs ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Get(destModel, selectFields, condition, conditionArgs...)
}

// 更新记录
func Update(model interface{}, fields any, condition interface{}, conditionArgs ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Update(model, fields, condition, conditionArgs...)
}

// 删除记录
func Delete(model interface{}, PrimaryKeyID interface{}, condition interface{}, conditionArgs ...interface{}) (int64, error) {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Delete(model, PrimaryKeyID, condition, conditionArgs...)
}

// 创建记录
func Create(model interface{}) (int64, error) {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Create(model)
}

// 执行原生sql语句
func SqlExec(sql interface{}, args ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.SqlExec(sql, args...)

}

// 原生sql查询
func SqlQuery(sql interface{}, args ...interface{}) (resultsSlice []map[string]interface{}, err error) {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.SqlQuery(sql, args...)
}

// 原生sql查询解析到结构体或MAP
func SqlQueryScan(destModel interface{}, sql interface{}, args ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.SqlQueryScan(destModel, sql, args...)
}

func QueryScan(destModel interface{}, extra interface{}, condition interface{}, args ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.QueryScan(destModel, extra, condition, args...)
}

func Count(model interface{}, intPrt *int64, condition interface{}, conditionArgs ...interface{}) error {
	ormEngine := packageObject.GetEngine(context.Background())
	return ormEngine.Count(model, intPrt, condition, conditionArgs...)
}

func ListScan(l *List, model interface{}, destModels interface{}) {
	ormEngine := packageObject.GetEngine(context.Background())
	ormEngine.ListScan(l, model, destModels)
}

func reflectSliceModelCall(destModels any, methodName string, session DbTransactionSession) {
	if methodName == "" {
		return
	}
	ref := reflect.ValueOf(destModels)

	for {
		if ref.Kind().String() == "ptr" {
			ref = ref.Elem()
		} else {
			break
		}
	}

	if ref.Kind().String() != "slice" {
		return
	}

	for i := 0; i < ref.Len(); i++ {
		if ref.Index(i).Kind().String() == "ptr" {
			reflectMethodCallByReflect(ref.Index(i), methodName, session)
		} else {
			reflectMethodCallByReflect(ref.Index(i).Addr(), methodName, session)
		}
	}
}

func reflectMethodCallByReflect(reflectValue reflect.Value, methodName string, session DbTransactionSession) error {
	reflectType := reflectValue.Type()
	method, ok := reflectType.MethodByName(methodName)
	if !ok {
		return nil
	}

	if method.Type.NumIn() == 2 { //有2个输入参数，判断输入参数是否为basedboperat.DbTransactionSession
		if strings.Contains(method.Type.In(1).Name(), "DbTransactionSession") {
			ret := method.Func.Call([]reflect.Value{reflectValue, reflect.ValueOf(session)})
			if len(ret) > 0 {
				if ret[0].Type().Kind().String() == "interface" {
					if ret[0].Type().Name() == "error" {
						if ret[0].Interface() == nil {
							return nil
						}
						return ret[0].Interface().(error)
					}
				}
			}
		}
	} else {
		ret := method.Func.Call([]reflect.Value{reflectValue})
		if len(ret) > 0 {
			if ret[0].Type().Kind().String() == "interface" {
				if ret[0].Type().Name() == "error" {
					if ret[0].Interface() == nil {
						return nil
					}
					return ret[0].Interface().(error)
				}
			}
		}
	}

	return nil
}

func reflectMethodCall(destModel interface{}, methodName string, session DbTransactionSession) error {
	reflectType := reflect.TypeOf(destModel)
	method, ok := reflectType.MethodByName(methodName)
	if ok {
		if method.Type.NumIn() == 2 { //有2个输入参数，判断输入参数是否为basedboperat.DbTransactionSession
			if strings.Contains(method.Type.In(1).Name(), "DbTransactionSession") {
				ret := method.Func.Call([]reflect.Value{reflect.ValueOf(destModel), reflect.ValueOf(session)})
				if len(ret) > 0 {
					if ret[0].Type().Kind().String() == "interface" {
						if ret[0].Type().Name() == "error" {
							if ret[0].Interface() == nil {
								return nil
							}
							return ret[0].Interface().(error)
						}
					}
				}
			}
		} else {
			ret := method.Func.Call([]reflect.Value{reflect.ValueOf(destModel)})
			if len(ret) > 0 {
				if ret[0].Type().Kind().String() == "interface" {
					if ret[0].Type().Name() == "error" {
						if ret[0].Interface() == nil {
							return nil
						}
						return ret[0].Interface().(error)
					}
				}
			}
		}
	}
	return nil
}

func WithContext(ctx context.Context) (DbTransactionSession, error) {
	return packageObject.NewEngineSession(ctx)
}

// 新建事务
func NewTransaction() (DbTransactionSession, error) {
	return packageObject.NewEngineSession(context.Background())
}

func Sync(beans ...any) {
	ormEngine := packageObject.GetEngine(context.Background())
	ormEngine.Sync(beans...)
}

func queryCache(destModel any, cacheKey string) bool {
	bufInterface, ok := queryCacheMap.Load(cacheKey)
	if !ok { //缓存不存在
		return false
	}
	buf := bufInterface.(bytes.Buffer)
	decoder := gob.NewDecoder(&buf)
	decoder.Decode(destModel)
	return true
}

// 根据模型获取缓存时间 ， 如模型没有绑定缓存时间方法， 返回全局时间
func getModelExpire(destModel any) int64 {
	reflectType := reflect.TypeOf(destModel)

	//指针类型过滤
	for {
		if reflectType.Kind().String() == "ptr" {
			reflectType = reflectType.Elem()
		} else {
			break
		}
	}

	//切片类型过滤
	if reflectType.Kind().String() == "slice" {
		reflectType = reflectType.Elem()
	}

	if reflectType.Kind() != reflect.Ptr {
		reflectType = reflect.PtrTo(reflectType)
	}

	method, ok := reflectType.MethodByName("CacheExpire")

	if !ok {
		//没有绑定到期函数，认定该对象不需要缓存 返回全局缓存时间
		return globalCacheExpire
	}

	ret := method.Func.Call([]reflect.Value{reflect.New(reflectType.Elem())})
	if len(ret) <= 0 {
		//没有返回值  ，认定该对象不需要缓存 返回全局缓存时间
		return globalCacheExpire
	}

	if ret[0].Type().Kind().String() != "int64" {
		//返回值不为int64 ，认定该对象不需要缓存 返回全局缓存时间
		return globalCacheExpire
	}
	return ret[0].Int()
}

func setCache(destModel any, cacheKey string, expire int64) bool {
	if expire == 0 {
		expire = getModelExpire(destModel)
	}
	if expire == 0 {
		return false
	}
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	enc.Encode(destModel)
	queryCacheMap.Store(cacheKey, buf)
	timer := time.NewTimer(time.Duration(expire) * time.Millisecond)
	go func(t *time.Timer, k string) {
		<-t.C
		queryCacheMap.Delete(k)
	}(timer, cacheKey)
	return true
}

func GenerateCacheKey(params ...any) string {
	var key string
	for _, v := range params {
		b, _ := json.Marshal(v)
		key = key + string(b)
	}
	key = system.MD5(key)
	return key
}

func ReflectStructID(obj interface{}) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("xorm")
		if tag == "pk" {
			return v.Field(i).Interface()
		}
	}

	field := v.FieldByName("ID")

	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

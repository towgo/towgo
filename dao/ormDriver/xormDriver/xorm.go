package xormDriver

import (
	"errors"
	"sync"
	"time"

	//_ "github.com/alexbrainman/odbc"
	//_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/towgo/towgo/os/tlog"
	"xorm.io/xorm"
	dblog "xorm.io/xorm/log"
	"xorm.io/xorm/names"
)

var masterDbs sync.Map
var slaveDbs sync.Map
var offlineSlaveDbs sync.Map
var offlineMasterDbs sync.Map
var dbid int64 = 0
var inited bool
var syncBeansChan chan interface{} = make(chan interface{}, 1)

type DsnConfig struct {
	DbType          string
	IsMaster        bool
	Dsn             string
	SqlLogLevel     int64
	SqlMaxIdleConns int64
	SqlMaxOpenConns int64
}

type Db struct {
	ID int64
	DsnConfig
	Engine *xorm.Engine
}

func Init() {
	if inited {
		return
	}
	inited = true
	checkDbHealthy()
	syncBeans()
}

func syncBeans() {
	go func() {
		for {
			bean := <-syncBeansChan
			db := DbMaster()
			if db == nil {
				continue
			}
			err := db.Engine.Sync2(bean)
			if err != nil {
				tlog.Errorf("sync beans {%+v} failed , %+v", bean, err)
				continue
			}
		}
	}()
}

func New(dsnConfigs []DsnConfig) {
	Init()
	for _, v := range dsnConfigs {

		var dbEngine *xorm.Engine
		var err error

		switch v.DbType {
		case "sqlite":
			dbEngine, err = xorm.NewEngine("sqlite3", v.Dsn)
		case "mysql":
			dbEngine, err = xorm.NewEngine("mysql", v.Dsn)
		case "postgres":
			dbEngine, err = xorm.NewEngine("postgres", v.Dsn)
		case "mssql":
			dbEngine, err = xorm.NewEngine("mssql", v.Dsn)
		default:
			err = errors.New("不支持的数据库类型")
		}

		if err != nil {
			tlog.Print(err.Error())
			continue
		}

		switch v.SqlLogLevel {
		case 5:
			dbEngine.Logger().SetLevel(dblog.LOG_UNKNOWN)
		case 4:
			dbEngine.Logger().SetLevel(dblog.LOG_INFO)
		case 3:
			dbEngine.Logger().SetLevel(dblog.LOG_ERR)
		case 2:
			dbEngine.Logger().SetLevel(dblog.LOG_WARNING)
		case 1:
			dbEngine.Logger().SetLevel(dblog.LOG_OFF)
		default:
			dbEngine.Logger().SetLevel(dblog.DEFAULT_LOG_LEVEL)
		}

		dbEngine.ShowSQL(true)

		if v.SqlMaxIdleConns == 0 {
			v.SqlMaxIdleConns = 10
		}
		if v.SqlMaxOpenConns == 0 {
			v.SqlMaxOpenConns = 10
		}
		dbEngine.SetMaxIdleConns(int(v.SqlMaxIdleConns))
		dbEngine.SetMaxOpenConns(int(v.SqlMaxOpenConns))

		dbEngine.SetMapper(names.GonicMapper{})
		dbEngine.SetConnMaxLifetime(time.Hour)
		dbEngine.SetConnMaxIdleTime(30 * time.Second) // 连接最大空闲时间（超时则关闭）
		dbEngine.SetConnMaxLifetime(1 * time.Hour)    // 连接最大存活时间（超时则重建）

		dbid++
		db := &Db{}
		db.DbType = v.DbType
		db.Dsn = v.Dsn
		db.IsMaster = v.IsMaster
		db.Engine = dbEngine
		if v.IsMaster {
			masterDbs.Store(dbid, db)
		} else {
			slaveDbs.Store(dbid, db)
		}
	}
}

// 健康检查
func checkDbHealthy() {
	go func() {
		for {
			masterDbs.Range(func(key, value any) bool {
				db := value.(*Db)
				err := db.Engine.Ping()
				if err != nil {
					masterDbs.Delete(key)
					offlineMasterDbs.Store(db.ID, db)
				}
				return true
			})

			slaveDbs.Range(func(key, value any) bool {
				db := value.(*Db)
				err := db.Engine.Ping()
				if err != nil {
					slaveDbs.Delete(key)
					offlineSlaveDbs.Store(db.ID, db)
				}
				return true
			})

			offlineMasterDbs.Range(func(key, value any) bool {
				db := value.(*Db)
				err := db.Engine.Ping()
				if err != nil {
					offlineMasterDbs.Delete(key)
					masterDbs.Store(db.ID, db)
				}
				return true
			})

			offlineSlaveDbs.Range(func(key, value any) bool {
				db := value.(*Db)
				err := db.Engine.Ping()
				if err != nil {
					offlineSlaveDbs.Delete(key)
					slaveDbs.Store(db.ID, db)
				}
				return true
			})

			time.Sleep(time.Second * 10)
		}
	}()
}

func Sync2(beans ...interface{}) {
	go func() {
		for _, v := range beans {
			syncBeansChan <- v
		}
	}()
}

// 随机获取主数据库
func DbMaster() *Db {
	var db *Db
	for {

		masterDbs.Range(func(_, value any) bool {
			db = value.(*Db)
			return false
		})
		if db == nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		break
	}
	return db
}

// 随机获取只读（备用）数据库
func DbSlave() *Db {
	var db *Db
	slaveDbs.Range(func(_, value any) bool {
		db = value.(*Db)
		return false
	})
	if db == nil {
		return DbMaster()
	}
	return db
}

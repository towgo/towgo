package gormDriver

import (
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm/schema"

	//"github.com/cengsin/oracle"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	retryCountLimit int64 = 100
)

var masterDbs sync.Map
var slaveDbs sync.Map
var syncBeansChan chan interface{} = make(chan interface{}, 1)

// var offlineDbs sync.Map
var dbid int64 = 0
var inited bool

type DsnConfig struct {
	DbType          string
	IsMaster        bool
	Dsn             string
	SqlLogLevel     int64
	SqlLogMode      string
	SqlMaxIdleConns int64
	SqlMaxOpenConns int64
}

type Db struct {
	ID int64
	DsnConfig
	Engine *gorm.DB
}

func Init() {
	if inited {
		return
	}
	inited = true
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
			db.Engine.AutoMigrate(bean)
		}
	}()
}

func New(dsnConfigs []DsnConfig) {
	Init()
	for _, v := range dsnConfigs {
		logLevel := logger.LogLevel(v.SqlLogLevel)
		switch strings.ToLower(v.SqlLogMode) {
		case "off":
			logLevel = logger.Silent
		case "error":
			logLevel = logger.Error
		case "info":
			logLevel = logger.Info
		case "debug":
			logLevel = logger.LogLevel(4)

		}
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
			logger.Config{
				SlowThreshold: time.Second, // 慢 SQL 阈值
				//LogLevel:                  logger.Error, // 日志级别
				LogLevel:                  logLevel, // 日志级别
				IgnoreRecordNotFoundError: true,     // 忽略ErrRecordNotFound（记录未找到）错误
				Colorful:                  false,    // 禁用彩色打印
			},
		)

		var dbEngine *gorm.DB
		var err error

		switch v.DbType {
		case "sqlite":
			dbEngine, err = gorm.Open(sqlite.Open(v.Dsn), &gorm.Config{
				Logger:         newLogger,
				NamingStrategy: schema.NamingStrategy{SingularTable: true},
			})
		case "mysql":
			dbEngine, err = gorm.Open(mysql.Open(v.Dsn), &gorm.Config{
				Logger:         newLogger,
				NamingStrategy: schema.NamingStrategy{SingularTable: true},
			})
		case "postgres":
			dbEngine, err = gorm.Open(postgres.Open(v.Dsn), &gorm.Config{
				Logger:         newLogger,
				NamingStrategy: schema.NamingStrategy{SingularTable: true},
			})
		case "mssql":
			dbEngine, err = gorm.Open(sqlserver.Open(v.Dsn), &gorm.Config{
				Logger:         newLogger,
				NamingStrategy: schema.NamingStrategy{SingularTable: true},
			})
		default:
			err = errors.New("不支持的数据库类型")
		}

		if err != nil {
			log.Print(err.Error())
			continue
		}

		if v.SqlMaxIdleConns == 0 {
			v.SqlMaxIdleConns = 10
		}
		if v.SqlMaxOpenConns == 0 {
			v.SqlMaxOpenConns = 10
		}
		sqlDB, err := dbEngine.DB()
		if err != nil {
			log.Print(err.Error())
			continue
		}
		sqlDB.SetMaxIdleConns(int(v.SqlMaxIdleConns))
		sqlDB.SetMaxOpenConns(int(v.SqlMaxOpenConns))
		sqlDB.SetConnMaxLifetime(time.Hour)
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

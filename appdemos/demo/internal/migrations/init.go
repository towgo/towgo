package migrations

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

var Migrations []interface{}
var gormEngine *gorm.DB

func init() {
	var err error
	dbConfig := g.DB().GetConfig()
	dsn := strings.Replace(dbConfig.Link, dbConfig.Type+":", "", 1) // 只替换第一个匹配项
	gormEngine, err = gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}
}
func AddMigrate(d ...interface{}) {
	Migrations = append(Migrations, d...)
}

func Sync(ctx context.Context) {
	err := gormEngine.AutoMigrate(Migrations...)
	if err != nil {
		glog.Error(ctx, err)
	}
	for _, m := range Migrations {
		err = gormEngine.AutoMigrate(m)
		if err != nil {
			glog.Error(ctx, err)
		}
	}
}

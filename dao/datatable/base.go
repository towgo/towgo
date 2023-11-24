package datatable

import (
	"sync"

	"github.com/towgo/towgo/dao/datatable/engines"
)

type TableOrmDriver struct {
	mode   string
	Engine engines.TableOrmEngine
}

var mu sync.Mutex
var ormEngine *TableOrmDriver

func newEngine(mode string) *TableOrmDriver {
	var engine engines.TableOrmEngine
	if mode == "xorm" {
		engine = &engines.Xorm{}
	} else if mode == "gorm" {
		engine = &engines.Gorm{}
	} else {

	}
	return &TableOrmDriver{mode: mode, Engine: engine}
}

func GetDriver() *TableOrmDriver {
	mu.Lock() // 如果实例存在没有必要加锁
	defer mu.Unlock()

	if ormEngine == nil {
		ormEngine = newEngine("xorm")
	}
	return ormEngine
}

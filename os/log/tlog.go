package log

import (
	"github.com/gogf/gf/v2/os/glog"
	"github.com/towgo/towgo/errors/terror"
	"github.com/towgo/towgo/os/tcfg"
)

func init() {
	data, err := tcfg.GetConfig().Get(ConfigNodeNameLogger)
	if err != nil {
		panic(terror.Wrap(err, "logger config init error"))
	}
	if data == nil {
		err = logger.SetConfig(glog.DefaultConfig())
		if err != nil {
			panic(terror.Wrap(err, "logger config init error"))
		}
		logger.SetTimeFormat("2006-01-02 15:04:05")
		err = logger.SetPath(DefaultLogPath)
		if err != nil {
			panic(terror.Wrap(err, "logger config init error"))
		}
	} else {
		err = logger.SetConfigWithMap(data.(map[string]interface{}))
		if err != nil {
			panic(terror.Wrap(err, "logger config init error"))
		}
	}

}

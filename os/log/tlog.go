package log

import (
	"github.com/towgo/towgo/errors/terror"
	"github.com/towgo/towgo/os/tcfg"
)

func init() {
	data, err := tcfg.GetConfig().Get(ConfigNodeNameLogger)
	if err != nil {
		panic(terror.Wrap(err, "logger config init error"))
	}
	err = logger.SetConfigWithMap(data.(map[string]interface{}))
	if err != nil {
		panic(terror.Wrap(err, "logger config init error"))
	}
}

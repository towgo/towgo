package log

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
)

var ConfigNodeNameLogger = "logger"
var DefaultLogPath = "glog"
var glogConfig glog.Config
var logger *glog.Logger = glog.New()
var ctx context.Context = context.TODO()

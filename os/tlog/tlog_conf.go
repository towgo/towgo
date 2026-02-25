package tlog

import (
	"context"
	"github.com/gogf/gf/v2/os/glog"
)

var ConfigNodeNameLogger = "logger"
var DefaultLogPath = "logs"
var glogConfig glog.Config
var logger *glog.Logger = glog.New()
var ctx context.Context = context.TODO()

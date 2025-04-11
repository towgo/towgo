package log

func init() {
	logger.SetTimeFormat("2006-01-02 15:04:05")
	logger.Path(DefaultLogPath)
}
func SetConfig(config map[string]interface{}) {
	err := logger.SetConfigWithMap(config)
	if err != nil {
		panic(err)
	}
}
func Println(v ...interface{}) {
	Print(v...)
}
func Print(v ...interface{}) {
	logger.Print(ctx, v...)
}
func Printf(format string, v ...interface{}) {
	logger.Printf(ctx, format, v...)
}

// Fatal prints the logging content with [FATA] header and newline, then exit the current process.
func Fatal(v ...interface{}) {
	logger.Fatal(ctx, v...)
}

// Fatalf prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(ctx, format, v...)
}

// Panic prints the logging content with [PANI] header and newline, then panics.
func Panic(v ...interface{}) {
	logger.Panic(ctx, v...)
}

// Panicf prints the logging content with [PANI] header, custom format and newline, then panics.
func Panicf(format string, v ...interface{}) {
	logger.Panicf(ctx, format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func Info(v ...interface{}) {
	logger.Info(ctx, v...)
}

// Infof prints the logging content with [INFO] header, custom format and newline.
func Infof(format string, v ...interface{}) {
	logger.Infof(ctx, format, v...)
}

// Debug prints the logging content with [DEBU] header and newline.
func Debug(v ...interface{}) {
	logger.Debug(ctx, v...)
}

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(format string, v ...interface{}) {
	logger.Debugf(ctx, format, v...)

}

// Notice prints the logging content with [NOTI] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Notice(v ...interface{}) {
	logger.Notice(ctx, v...)
}

// Noticef prints the logging content with [NOTI] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Noticef(format string, v ...interface{}) {
	logger.Noticef(ctx, format, v...)

}

// Warning prints the logging content with [WARN] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Warning(v ...interface{}) {
	logger.Warning(ctx, v...)
}

// Warningf prints the logging content with [WARN] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Warningf(format string, v ...interface{}) {
	logger.Warningf(ctx, format, v...)
}

// Error prints the logging content with [ERRO] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Error(v ...interface{}) {
	logger.Error(ctx, v...)
}

// Errorf prints the logging content with [ERRO] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Errorf(format string, v ...interface{}) {
	logger.Errorf(ctx, format, v...)
}

// Critical prints the logging content with [CRIT] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Critical(v ...interface{}) {
	logger.Critical(ctx, v...)
}

// Criticalf prints the logging content with [CRIT] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Criticalf(format string, v ...interface{}) {
	logger.Criticalf(ctx, format, v...)
}

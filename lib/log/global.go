package log

var Global Logger

func init() {
	Global = NewLogger()
}

func AddFilter(name string, lvl Level, writer LogWriter) {
	Global.AddFilter(name, lvl, writer)
}

func Finest(arg0 interface{}, args ...interface{}) {
	Global.Finest(arg0, args...)
}

func Fine(arg0 interface{}, args ...interface{}) {
	Global.Fine(arg0, args...)
}

func Debug(arg0 interface{}, args ...interface{}) {
	Global.Debug(arg0, args...)
}

func Trace(arg0 interface{}, args ...interface{}) {
	Global.Trace(arg0, args...)
}

func Info(arg0 interface{}, args ...interface{}) {
	Global.Info(arg0, args...)
}

func Warn(arg0 interface{}, args ...interface{}) error {
	return Global.Warn(arg0, args...)
}

func Error(arg0 interface{}, args ...interface{}) error {
	return Global.Error(arg0, args...)
}

func Critical(arg0 interface{}, args ...interface{}) error {
	return Global.Critical(arg0, args...)
}

func Logf(lvl Level, format string, args ...interface{}) {
	Global.Logf(lvl, format, args...)
}

func Close() {
	Global.Close()
}

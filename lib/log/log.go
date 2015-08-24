package log

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

type level int

const (
	FINEST level = iota
	FINE
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	CRITICAL
)

var levelStrings = [...]string{"FNST", "FINE", "DEBG", "TRAC", "INFO", "WARN",
	"EROR", "CRIT"}

func (l level) String() string {
	if l < 0 || int(l) > len(levelStrings) {
		return "UNKNOWN"
	}
	return levelStrings[int(l)]
}

type LogRecord struct {
	Level   level
	Created time.Time
	Source  string
	Message string
}

var (
	LogBufLen = 32
)

type LogWriter interface {
	LogWrite(rec *LogRecord)
	Close()
}

type Filter struct {
	Level level
	LogWriter
}

type Logger map[string]*Filter

func NewDefaultLogger(lvl level) Logger {
	return Logger{"stdout": &Filter{lvl, NewConsoleLogWriter()}}
}

func (log Logger) AddFilter(name string, lvl level, writer LogWriter) Logger {
	log[name] = &Filter{lvl, writer}
	return log
}

func (log Logger) Close() {
	for name, filter := range log {
		filter.Close()
		delete(log, name)
	}
}

// Send a formatted log message internally
func (log Logger) intLogf(lvl level, format string, args ...interface{}) {
	skip := true
	for _, filter := range log {
		if lvl >= filter.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	// Determine caller func
	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprintf("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: msg,
	}

	// Dispatch the logs
	for _, filter := range log {
		if lvl < filter.Level {
			continue
		}
		filter.LogWrite(rec)
	}
}

// Send a closure log message internally
func (log Logger) intLogc(lvl level, closure func() string) {
	skip := true
	for _, filter := range log {
		if lvl >= filter.Level {
			skip = false
			break
		}
	}
	if skip {
		return
	}

	// Determine caller func
	pc, _, lineno, ok := runtime.Caller(2)
	src := ""
	if ok {
		src = fmt.Sprint("%s:%d", runtime.FuncForPC(pc).Name(), lineno)
	}

	// Make the log record
	rec := &LogRecord{
		Level:   lvl,
		Created: time.Now(),
		Source:  src,
		Message: closure(),
	}

	// Dispatch the logs
	for _, filter := range log {
		if lvl < filter.Level {
			continue
		}
		filter.LogWrite(rec)
	}
}

func (log Logger) Finest(arg0 interface{}, args ...interface{}) {
	const lvl = FINEST
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Fine(arg0 interface{}, args ...interface{}) {
	const lvl = FINE
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Debug is a utility method for debug log message
// The behavior of Debug depends on the first argument:
// - arg0 is string
//   When given a string as the first argument, it is interpreted as a format
//   for the latter arguments.
// - arg0 is a func() string
//   When given a closure of type func() sting, this logs the string returned by
//   the closure if it will be logged. The closure runs at most one time.
// - arg0 is interface{}
//   When given anything else, the log message will be each of the arguments
//   formatted with %v and separated by spaces (ala Sprint).
func (log Logger) Debug(arg0 interface{}, args ...interface{}) {
	const lvl = DEBUG
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Trace(arg0 interface{}, args ...interface{}) {
	const lvl = TRACE
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func (log Logger) Info(arg0 interface{}, args ...interface{}) {
	const lvl = INFO
	switch first := arg0.(type) {
	case string:
		log.intLogf(lvl, first, args...)
	case func() string:
		log.intLogc(lvl, first)
	default:
		log.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// At the warning level and higher, there is no performance benefit if the
// message is not actually logged, because all formats are processed and all
// closure are executed to format the error message
func (log Logger) Warn(arg0 interface{}, args ...interface{}) error {
	const lvl = WARNING
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Error(arg0 interface{}, args ...interface{}) error {
	const lvl = ERROR
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

func (log Logger) Critical(arg0 interface{}, args ...interface{}) error {
	const lvl = CRITICAL
	var msg string
	switch first := arg0.(type) {
	case string:
		msg = fmt.Sprintf(first, args...)
	case func() string:
		msg = first()
	default:
		msg = fmt.Sprintf(fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
	log.intLogf(lvl, msg)
	return errors.New(msg)
}

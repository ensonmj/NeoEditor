package log

import (
	"fmt"
	"os"
	"time"
)

type ConsoleLogWriter struct {
	format string
	rec    chan *LogRecord
}

func NewConsoleLogWriter() *ConsoleLogWriter {
	w := &ConsoleLogWriter{
		format: "[%T %D][%L][%s]%M",
		rec:    make(chan *LogRecord, LogBufLen),
	}
	go func() {
		for rec := range w.rec {
			fmt.Fprint(os.Stdout, formatLogRecord(w.format, rec))
		}
	}()

	return w
}

func (w *ConsoleLogWriter) LogWrite(rec *LogRecord) {
	w.rec <- rec
}

func (w *ConsoleLogWriter) Close() {
	close(w.rec)
	time.Sleep(50 * time.Millisecond) // Try to give console I/O time to complete
}

func (w *ConsoleLogWriter) SetFormat(format string) *ConsoleLogWriter {
	w.format = format
	return w
}

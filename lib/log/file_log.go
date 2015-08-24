package log

import (
	"fmt"
	"os"
	"time"
)

type FileLogWriter struct {
	fPath          string
	file           *os.File
	format         string
	rec            chan *LogRecord
	header, footer string
}

func NewFileLogWriter(filePath string) *FileLogWriter {
	w := &FileLogWriter{
		fPath:  filePath,
		format: "[%D %T][%L][%s]%M",
		rec:    make(chan *LogRecord, LogBufLen),
		header: "==========*** This line is the header of the log ***==========",
		footer: "==========*** This line is the footer of the log ***==========",
	}

	fd, err := os.OpenFile(w.fPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil
	}
	w.file = fd

	now := time.Now()
	fmt.Fprintf(w.file, formatLogRecord(w.format, &LogRecord{Created: now,
		Message: w.header}))

	go func() {
		defer func() {
			if w.file != nil {
				fmt.Fprintf(w.file, formatLogRecord(w.format,
					&LogRecord{Created: time.Now(), Message: w.footer}))
				w.file.Close()
			}
		}()

		for {
			select {
			case rec, ok := <-w.rec:
				if !ok {
					return
				}
				_, err := fmt.Fprintf(w.file, formatLogRecord(w.format, rec))
				if err != nil {
					fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", w.fPath, err)
					return
				}
			}
		}
	}()

	return w
}

func (w *FileLogWriter) LogWrite(rec *LogRecord) {
	w.rec <- rec
}

func (w *FileLogWriter) Close() {
	close(w.rec)
	w.file.Sync()
}

func (w *FileLogWriter) SetFormat(format string) *FileLogWriter {
	w.format = format
	return w
}

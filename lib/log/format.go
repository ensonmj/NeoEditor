package log

import (
	"bytes"
	"fmt"
	"strings"
)

type formatCacheT struct {
	LastUpdateSec        int64
	shortTime, shortDate string
	longTime, longDate   string
}

var formatCache = &formatCacheT{}

// Known format codes
// %T - Time (23:24:25 CST)
// %t - Time (23:24)
// %D - Date (2015/08/24)
// %d - Date (08/24/15)
// %L - Level (FNST, FINE, DEBG, TRAC, WARN, EROR, CRIT)
// %S - Source
// %s - FileName
// %M - Message
// Ignore unknown formats
// Recomanded: "[%D %T][%L][%S]%M"
func formatLogRecord(format string, rec *LogRecord) string {
	if rec == nil {
		return "<nil>"
	}

	if len(format) == 0 {
		return ""
	}

	sec := rec.Created.UnixNano() / 1e9
	cache := *formatCache
	if cache.LastUpdateSec != sec {
		month, day, year := rec.Created.Month(), rec.Created.Day(), rec.Created.Year()
		hour, minute, second := rec.Created.Hour(), rec.Created.Minute(), rec.Created.Second()
		zone, _ := rec.Created.Zone()
		updated := &formatCacheT{
			LastUpdateSec: sec,
			shortTime:     fmt.Sprintf("%02d:%02d", hour, minute),
			shortDate:     fmt.Sprintf("%02d/%02d/%02d", day, month, year%100),
			longTime:      fmt.Sprintf("%02d:%02d:%02d %s", hour, minute, second, zone),
			longDate:      fmt.Sprintf("%04d/%02d/%02d", year, month, day),
		}
		cache = *updated
		formatCache = updated
	}

	// split the string into pieces by % signs
	pieces := bytes.Split([]byte(format), []byte{'%'})

	out := bytes.NewBuffer(make([]byte, 0, 64))
	// iterate over the pieces, replacing known formats
	for i, piece := range pieces {
		if i > 0 && len(piece) > 0 {
			switch piece[0] {
			case 'T':
				out.WriteString(cache.longTime)
			case 't':
				out.WriteString(cache.shortTime)
			case 'D':
				out.WriteString(cache.longDate)
			case 'd':
				out.WriteString(cache.shortDate)
			case 'L':
				out.WriteString(levelStrings[rec.Level])
			case 'S':
				out.WriteString(rec.Source)
			case 's':
				slice := strings.Split(rec.Source, "/")
				out.WriteString(slice[len(slice)-1])
			case 'M':
				out.WriteString(rec.Message)
			}
			if len(piece) > 1 {
				out.Write(piece[1:])
			}
		} else if len(piece) > 0 {
			out.Write(piece)
		}
	}
	out.WriteByte('\n')

	return out.String()
}

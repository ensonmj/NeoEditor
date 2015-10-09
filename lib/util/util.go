package util

import "runtime"

func StackTrace(all bool) string {
	buf := make([]byte, 1024)

	for {
		size := runtime.Stack(buf, all)
		// the size of the buffer may be not enough to hold the stacktrace,
		// so double the buffer size
		if size == len(buf) {
			buf = make([]byte, len(buf)<<1)
			continue
		}

		return string(buf[:size])
	}
}

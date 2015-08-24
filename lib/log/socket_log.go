package log

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// This log write sends output to a socket
type SocketLogWriter chan *LogRecord

func NewSocketLogWriter(proto, hostport string) SocketLogWriter {
	sock, err := net.Dial(proto, hostport)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewFileLogWriter(%q); %s\n", hostport, err)
		return nil
	}

	w := SocketLogWriter(make(chan *LogRecord, LogBufLen))

	go func() {
		defer func() {
			if sock != nil && proto == "tcp" {
				sock.Close()
			}
		}()

		for rec := range w {
			// Marshall into JSON
			js, err := json.Marshal(rec)
			if err != nil {
				fmt.Fprintf(os.Stderr, "NewFileLogWriter(%q); %s\n", hostport, err)
				return
			}

			_, err = sock.Write(js)
			if err != nil {
				fmt.Fprintf(os.Stderr, "NewFileLogWriter(%q); %s\n", hostport, err)
				return
			}
		}
	}()

	return w
}

func (w SocketLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

func (w SocketLogWriter) Close() {
	close(w)
}

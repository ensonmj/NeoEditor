package neoeditor

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
	zmq "github.com/pebbe/zmq4"
)

const (
	chanBufLen = 16
)

type Editor struct {
	//cmds      chan string
	events    chan codec.Envelope
	done      chan bool
	pm        plugin.PluginManager
	mode      Mode
	bufs      []*Buffer
	activeBuf int
	tabs      []*Tab
	activeTab int
}

// Command line flags
var (
	showDebug = flag.Bool("debug", false, "Display debug log")
)

func NewEditor() (*Editor, error) {
	log.AddFilter("file", log.DEBUG, log.NewFileLogWriter("./ned.log"))

	// profile
	go func() {
		http.ListenAndServe("127.0.0.1:5197", nil)
	}()

	ed := &Editor{mode: Normal, pm: make(plugin.PluginManager, 1)}

	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	//ed.cmds = make(chan string, chanBufLen)
	ed.events = make(chan codec.Envelope, chanBufLen)
	ed.done = make(chan bool)

	rep, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		return nil, err
	}
	rep.Bind("inproc://command")
	rep.Bind("tcp://*:5198")

	// monitor request
	go func() {
		for {
			cmd, err := rep.Recv(0)
			log.Debug("received:%v,%v", cmd, err)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			ed.DispatchCommand(string(cmd))
		}
	}()

	// the publisher need to sleep a litter before starting to publish
	pub, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	pub.Bind("inproc://notification")
	pub.Bind("tcp://*:5199")

	// broadcast notification
	go func() {
		for {
			ev := <-ed.events

			// env.Method as topic, and env.Arguments as content
			topic := fmt.Sprintf("%s", ev.Method)
			msg, _ := codec.Serialize(ev.Arguments)
			log.Debug("broadcast event:%s%s", topic, string(msg))
			pub.Send(topic, zmq.SNDMORE)
			pub.Send(string(msg), 0)
		}
	}()

	flag.Parse()
	// Assuming that all extra arguments are files
	if files := flag.Args(); len(files) > 0 {
		for _, fPath := range files {
			buf, err := NewBuffer(fPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				log.Warn("open file[%s] err:%s", fPath, err)
				continue
			}
			ed.bufs = append(ed.bufs, buf)
		}
	} else {
		// create a scratch buffer
		buf, _ := NewBuffer("", os.O_RDWR|os.O_CREATE, 0644)
		ed.bufs = append(ed.bufs, buf)
	}
	ed.activeBuf = 0
	b := ed.bufs[ed.activeBuf]
	v := b.View
	v.Contents = b.data
	log.Debug("View:%v", v)
	ed.PubEvent("updateView", v)

	// main loop
	go func() {
		for {
			select {
			//case cmd := <-ed.cmds:
			//ed.DispatchCommand(cmd)
			case <-ed.done:
				log.Debug("editor backend main loop exit")
				return
			}
		}
	}()

	return ed, nil
}

func (ed *Editor) ActiveTab() *Tab {
	return ed.tabs[ed.activeTab]
}

func (ed *Editor) ActiveWnd() *Window {
	t := ed.tabs[ed.activeTab]
	return t.Wnds[t.ActiveWnd]
}

func (ed *Editor) ActiveBuf() *Buffer {
	return ed.bufs[ed.activeBuf]
}

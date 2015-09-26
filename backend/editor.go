package neoeditor

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
	zmq "github.com/pebbe/zmq4"
)

const (
	chanBufLen = 16
)

type Editor struct {
	done      chan bool
	pm        plugin.PluginManager
	mode      Mode
	keys      []rune
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

	ed := &Editor{
		mode: Normal,
		pm:   make(plugin.PluginManager),
	}

	//ed.cmds = make(chan string, chanBufLen)
	ed.done = make(chan bool)
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	initEvent()
	registerCommands()
	registerModeAction()

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
				log.Debug("command monitor got an err:%v", err)
				close(ed.done)
				return
			}
			exit, err := dispatchCommand(ed, string(cmd))
			if exit {
				rep.Close()
				log.Debug("command monitor exit")
				return
			}
			if err != nil {
				pubEvent("error", err)
			}
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
			select {
			case ev := <-pollEvent():
				// env.Method as topic, and env.Arguments as content
				topic := fmt.Sprintf("%s", ev.Method)
				msg, _ := codec.Serialize(ev.Arguments)
				log.Debug("broadcast event:%s%s", topic, string(msg))
				pub.Send(topic, zmq.SNDMORE)
				pub.Send(string(msg), 0)
			case <-ed.done:
				pub.Send("exit", zmq.SNDMORE)
				pub.Send("neoeditor exit", 0)
				pub.Close()
				log.Debug("editor close notification broadcaster")
				return
			}
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
	pubEvent("updateView", v)

	// TODO: start ticker on demand
	ticker := time.NewTicker(5 * time.Millisecond)
	// main loop
	go func() {
		for {
			select {
			case <-ed.done:
				log.Debug("editor backend main loop exit")
				return
			case <-ticker.C:
				ed.ClearKeys()
			}
		}
	}()

	return ed, nil
}

func (ed *Editor) ActiveTab() *Tab {
	return ed.tabs[ed.activeTab]
}

func (ed *Editor) ActiveWnd() *Window {
	t := ed.ActiveTab()
	return t.Wnds[t.ActiveWnd]
}

func (ed *Editor) ActiveBuf() *Buffer {
	return ed.bufs[ed.activeBuf]
}

func (ed *Editor) ActiveView() *View {
	return &ed.ActiveBuf().View
}

func (ed *Editor) AccumulateKey(kp key.KeyPress) string {
	ed.keys = append(ed.keys, rune(kp.Key))
	return string(ed.keys)
}

func (ed *Editor) AllKeys() string {
	return string(ed.keys)
}

func (ed *Editor) ClearKeys() {
	ed.keys = nil
}

package neoeditor

import (
	"flag"
	"fmt"
	"os"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
	"github.com/ensonmj/NeoEditor/lib/util"
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
		mode: MNormal,
		pm:   make(plugin.PluginManager),
	}

	//ed.cmds = make(chan string, chanBufLen)
	ed.done = make(chan bool)
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	initEvent()
	registerCommands()
	registerModeAction()

	rep, err := zmq.NewSocket(zmq.REP)
	if err != nil {
		return nil, err
	}
	rep.Bind("inproc://register")
	rep.Bind("tcp://*5197")
	// monitor ui register
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Critical("%s, %s", err, util.StackTrace(false))
				panic(err)
			}
		}()
		for {
			// TODO: exit gracefully
			reqMsg, err := rep.Recv(0)
			log.Debug("receive ui register:%s,%s", reqMsg, err)
			if err != nil {
				log.Debug("register monitor got an err:%s", err)
				close(ed.done)
				return
			}
			var ui UI
			if err = codec.Deserialize([]byte(reqMsg), &ui); err != nil {
				log.Critical(err)
				close(ed.done)
				return
			}
			registerUI(&ui)
			ed.ActiveBuf().updateView()

			repMsg, _ := codec.Serialize(ui)
			rep.Send(string(repMsg), 0)
		}
	}()

	pull, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		return nil, err
	}
	pull.Bind("inproc://command")
	pull.Bind("tcp://*:5198")
	// monitor request
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Critical("%s, %s", err, util.StackTrace(false))
				panic(err)
			}
		}()
		for {
			cmd, err := pull.Recv(0)
			log.Debug("received:%s,%s", cmd, err)
			if err != nil {
				log.Debug("command monitor got an err:%s", err)
				close(ed.done)
				return
			}
			exit, err := dispatchCommand(ed, string(cmd))
			if exit {
				pull.Close()
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
		defer func() {
			if err := recover(); err != nil {
				log.Critical("%s, %s", err, util.StackTrace(false))
				panic(err)
			}
		}()
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

	log.Debug("view:%v", b.View)
	pubEvent("updateView", b.View)

	// TODO: start ticker on demand
	// main loop
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Critical("%s, %s", err, util.StackTrace(false))
				panic(err)
			}
		}()
		for {
			select {
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

func (ed *Editor) ClearKeys() string {
	str := string(ed.keys)
	ed.keys = nil
	return str
}

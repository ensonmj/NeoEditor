package neoeditor

import (
	"fmt"
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

type Mode int

const (
	Normal Mode = iota
	Insert
	Visual
)

func (m Mode) String() string {
	switch m {
	case Normal:
		return "Normal"
	case Insert:
		return "Insert"
	case Visual:
		return "Visual"
	default:
		return "Unknown"
	}
}

type Editor struct {
	//cmds      chan string
	events            chan codec.Envelope
	done              chan bool
	pm                plugin.PluginManager
	mode              Mode
	tabs              []*Tab
	activeTab         int
	bufs              []*Buffer
	activeBuf         int
	uiWidth, uiHeight int // active ui window size
}

func NewEditor() (*Editor, error) {
	log.AddFilter("file", log.DEBUG, log.NewFileLogWriter("./ned.log"))
	ed := &Editor{mode: Normal, pm: make(plugin.PluginManager, 1)}
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	//ed.cmds = make(chan string, chanBufLen)
	ed.events = make(chan codec.Envelope, chanBufLen)
	ed.done = make(chan bool)

	// create a scratch buffer
	buf, _ := NewBuffer("", os.O_RDWR|os.O_CREATE, 0644)
	ed.bufs = append(ed.bufs, buf)
	ed.activeBuf = 0

	rep, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		return nil, err
	}
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
	//pub.Bind("tcp://*:5199")
	pub.Bind("inproc://notification")

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

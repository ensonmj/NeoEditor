package neoeditor

import (
	//"fmt"
	"os"
	"time"

	"github.com/ensonmj/NeoEditor/lib/codec"
	//"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
	zmq "github.com/pebbe/zmq4"
)

const (
	chanBufLen = 16
)

type Editor struct {
	//kps chan key.KeyPress
	//cmds      chan string
	events    chan codec.Envelope
	done      chan bool
	pm        plugin.PluginManager
	tabs      []*Tab
	activeTab int
	bufs      []*Buffer
	activeBuf int
}

func NewEditor() (*Editor, error) {
	log.AddFilter("file", log.DEBUG, log.NewFileLogWriter("./ned.log"))
	ed := &Editor{pm: make(plugin.PluginManager, 1)}
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	//ed.kps = make(chan key.KeyPress, chanBufLen)
	//ed.cmds = make(chan string, chanBufLen)
	ed.events = make(chan codec.Envelope, chanBufLen)
	ed.done = make(chan bool)

	buf, _ := NewBuffer("buf.txt", os.O_RDWR|os.O_CREATE, 0644)
	ed.bufs = append(ed.bufs, buf)
	ed.activeBuf = 0

	rep, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		return nil, err
	}
	rep.Bind("tcp://*:5198")

	// the publisher need to sleep a litter before starting to publish
	pub, err := zmq.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	pub.Bind("tcp://*:5199")

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

	// broadcast notification
	go func() {
		for {
			ev := <-ed.events

			// env.Method as topic, and env.Arguments as content
			//topic := fmt.Sprintf("%s ", ev.Method)
			topic := "1 "
			pub.Send(topic, zmq.SNDMORE)
			msg, _ := codec.Serialize(ev.Arguments)
			log.Debug("broadcast event:%s%s", topic, string(msg))
			pub.Send(string(msg), 0)
		}
	}()

	// main loop
	go func() {
		for {
			select {
			//case cmd := <-ed.cmds:
			//ed.DispatchCommand(cmd)
			//case kp := <-ed.kps:
			//ed.handleKeyPress(kp)
			case <-ed.done:
				log.Debug("editor backend main loop exit")
				return
			}
		}
	}()

	return ed, nil
}

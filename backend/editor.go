package neoeditor

import (
	"os"

	"github.com/ensonmj/NeoEditor/backend/events"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
)

var Ned *Editor

func init() {
	Ned = NewEditor()
}

type Editor struct {
	pm        plugin.PluginManager
	tabs      []*Tab
	activeTab int
	bufs      []*Buffer
	activeBuf int
	chars     chan rune
	events    map[string]events.Event
}

func NewEditor() *Editor {
	log.AddFilter("file", log.DEBUG, log.NewFileLogWriter("./ned.log"))
	ed := &Editor{pm: make(plugin.PluginManager, 1), chars: make(chan rune, 32)}
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	buf, _ := NewBuffer("buf.txt", os.O_RDWR|os.O_CREATE, 0644)
	ed.bufs = append(ed.bufs, buf)
	ed.events = make(map[string]events.Event)
	ed.RegisterPublisher("bufferChanged", &events.BufferChanged{})
	go func() {
		for {
			select {
			case char := <-ed.chars:
				chars := make([]rune, 0, 1)
				chars = append(chars, char)
				ed.bufs[ed.activeBuf].Append(chars)

				ed.NotifyEvent("bufferChanged", ed.bufs[ed.activeBuf].Contents())
			}
		}
	}()

	return ed
}

func (ed *Editor) RegisterPublisher(event string, pub events.Event) {
	ed.events[event] = pub
}

func (ed *Editor) RegisterListener(event string, l events.Listener) {
	ed.events[event].AddListener(l)

}

func (ed *Editor) NotifyEvent(event string, args ...interface{}) {
	ed.events[event].Notify(args...)
}

func (ed *Editor) HandleInput(kp key.KeyPress) {
	log.Debug("receive key press:%v", kp)
	if kp.Ctrl && kp.Key == 's' {
		log.Debug("save buffer:%s", ed.bufs[ed.activeBuf])
		ed.bufs[ed.activeBuf].Close()
		return
	}

	ed.chars <- rune(kp.Key)
}

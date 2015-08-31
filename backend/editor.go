package neoeditor

import (
	"os"

	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
)

type Editor struct {
	pm        plugin.PluginManager
	tabs      []*Tab
	activeTab int
	bufs      []*Buffer
	activeBuf int
	chars     chan rune
}

func NewEditor() *Editor {
	log.AddFilter("backend", log.DEBUG, log.NewFileLogWriter("./neoeditor.log"))
	ed := &Editor{pm: make(plugin.PluginManager, 1), chars: make(chan rune, 32)}
	xui := &plugin.DummyPlugin{}
	xui.Register(ed.pm)

	buf, _ := NewBuffer("buf.txt", os.O_RDWR|os.O_CREATE, 0644)
	ed.bufs = append(ed.bufs, buf)
	go func() {
		for {
			select {
			case char := <-ed.chars:
				chars := make([]rune, 0, 1)
				chars = append(chars, char)
				ed.bufs[ed.activeBuf].Append(chars)
			}
		}
	}()

	return ed
}

func (ed *Editor) HandleInput(kp KeyPress) {
	log.Debug("receive key press:%v", kp)
	if kp.Ctrl && kp.Key == 's' {
		log.Debug("save buffer:%s", ed.bufs[ed.activeBuf])
		ed.bufs[ed.activeBuf].Close()
		return
	}

	ed.chars <- rune(kp.Key)
}

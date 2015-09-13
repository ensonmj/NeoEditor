package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
)

func (ed *Editor) DispatchCommand(cmd string) {
	log.Debug("receive command:%s", cmd)
	var payload codec.RawMessage
	env := codec.Envelope{
		Arguments: &payload,
	}
	if err := codec.Deserialize([]byte(cmd), &env); err != nil {
		log.Critical(err)
		return
	}
	log.Debug("parse command:{%s, %v}", env.Method, payload)

	switch env.Method {
	case "KeyPress":
		log.Debug("receive command [KeyPress]")
		var kp key.KeyPress
		if err := codec.Deserialize(payload, &kp); err != nil {
			log.Critical(err)
			return
		}
		log.Debug("parse command [KeyPress] arguments:%v", kp)
		ed.handleKeyPress(kp)
	}
}

func (ed *Editor) handleKeyPress(kp key.KeyPress) {
	log.Debug("receive key press:%v", kp)
	if kp.Ctrl && kp.Key == 'q' {
		close(ed.done)
		return
	}
	if kp.Ctrl && kp.Key == 's' {
		log.Debug("save buffer:%s", ed.bufs[ed.activeBuf])
		ed.bufs[ed.activeBuf].Close()
		return
	}

	// parse keypress

	ed.bufs[ed.activeBuf].Append([]rune(kp.Text))
	ed.PubEvent("updateView", ed.bufs[ed.activeBuf].Contents())
}

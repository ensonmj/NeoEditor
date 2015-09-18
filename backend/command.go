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
	ed.ResolvMode(kp)
}

func (ed *Editor) ResolvMode(kp key.KeyPress) {
	if kp.Key == key.Escape {
		ed.mode = Normal
		log.Debug("change mode to:%s", ed.mode)
		return
	}

	switch ed.mode {
	case Normal:
		if kp.Key == 'i' {
			ed.mode = Insert
			log.Debug("change mode to:%s", ed.mode)
		}
	case Insert:
		switch kp.Key {
		case key.Left:
			cmd := CmdMoveCursor{Direction: CLeft, Repeat: 1}
			cmd.Run(ed)
		case key.Up:
			cmd := CmdMoveCursor{Direction: CUp, Repeat: 1}
			cmd.Run(ed)
		case key.Right:
			cmd := CmdMoveCursor{Direction: CRight, Repeat: 1}
			cmd.Run(ed)
		case key.Down:
			cmd := CmdMoveCursor{Direction: CDown, Repeat: 1}
			cmd.Run(ed)
		default:
			cmd := CmdInsertRune{data: string(rune(kp.Key))}
			cmd.Run(ed)
		}
	case Visual:
	default:
	}
}

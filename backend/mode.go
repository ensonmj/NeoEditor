package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
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
		// test split
		if kp.Key == 's' {
			log.Debug("split window")
			ed.ActiveWnd().Split(Horizontal)
			ed.PubEvent("updateWnd", ed.ActiveWnd())
		}
	case Insert:
		switch kp.Key {
		case key.Left:
			cmd := CmdMoveCursor{Direction: Left, Repeat: 1}
			cmd.Run(ed)
		case key.Up:
			cmd := CmdMoveCursor{Direction: Up, Repeat: 1}
			cmd.Run(ed)
		case key.Right:
			cmd := CmdMoveCursor{Direction: Right, Repeat: 1}
			cmd.Run(ed)
		case key.Down:
			cmd := CmdMoveCursor{Direction: Down, Repeat: 1}
			cmd.Run(ed)
		default:
			cmd := CmdInsertRune{data: string(rune(kp.Key))}
			cmd.Run(ed)
		}
	case Visual:
	default:
	}
}

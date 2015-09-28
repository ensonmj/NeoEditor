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
	Command
)

// return true if program will exit
type KeyAction map[key.KeyPress]func(*Editor, key.KeyPress) (bool, error)

var modeActors = map[Mode]KeyAction{}

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

func registerModeAction() {
	nKA, iKA, vKA, cKA := KeyAction{}, KeyAction{}, KeyAction{}, KeyAction{}

	// normal
	nKA[key.KeyPress{Key: key.Left}] = moveCursor
	nKA[key.KeyPress{Key: key.Up}] = moveCursor
	nKA[key.KeyPress{Key: key.Right}] = moveCursor
	nKA[key.KeyPress{Key: key.Down}] = moveCursor
	nKA[key.KeyPress{Ctrl: true, Key: 'q'}] = func(*Editor, key.KeyPress) (bool, error) {
		return true, nil
	}

	// insert
	//iKA[key.KeyPress{Key: key.Escape}] = resolvMode

	// visual
	//vKA[key.KeyPress{Key: key.Escape}] = resolvMode

	// command
	cKA[key.KeyPress{Key: 'w'}] = func(ed *Editor, kp key.KeyPress) (bool, error) {
		ed.ActiveBuf().Save()
		return false, nil
	}

	modeActors[Normal] = nKA
	modeActors[Insert] = iKA
	modeActors[Visual] = vKA
	modeActors[Command] = cKA
}

// TODO: find action according to accumulated keys
func runModeAction(ed *Editor, kp key.KeyPress) (bool, error) {
	log.Debug("mode:%v kp:%v", ed.mode, kp)
	if resolvMode(ed, kp) {
		return false, nil
	}

	keyAction := modeActors[ed.mode]
	if actor, ok := keyAction[kp]; ok {
		return actor(ed, kp)
	} else {
		if ed.mode == Insert {
			//ed.ActiveBuf().Insert([]rune(string(kp.Key)))
			ed.ActiveBuf().Insert(rune(kp.Key))
		} else {
			var str string
			if kp.Key == key.Enter {
				str = ed.ClearKeys()
			} else {
				str = ed.AccumulateKey(kp)
			}
			log.Debug("accumulated keys:%s", str)
		}
	}

	return false, nil
}

func resolvMode(ed *Editor, kp key.KeyPress) bool {
	if kp.Key == key.Escape {
		ed.mode = Normal
		log.Debug("change mode to:%s", ed.mode)
		return true
	}

	changed := false
	switch ed.mode {
	case Normal:
		switch kp.Key {
		case 'i':
			ed.mode = Insert
			changed = true
		case ':':
			ed.mode = Command
			changed = true
		}
	case Insert:
	case Visual:
	}

	if changed {
		log.Debug("change mode to:%s", ed.mode)
		return true
	} else {
		log.Debug("keep mode in:%s", ed.mode)
		return false
	}
}

package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/gpmgo/gopm/modules/log"
)

type View struct {
	XOffset          int //must be 0 when wrap is on
	YOffset          int //the line displayed in the top of window
	RCursor, CCursor int //cursor position in the buffer(row,column)
	XCursor, YCursor int //cursor position in the screen
	//xUpdate, yUpdate int //start position for redraw
	Contents [][]rune
}

//type CmdMoveCursor struct {
//Direction
//Repeat int
//}

//func (c CmdMoveCursor) Run(ed *Editor) error {
//v := ed.ActiveView()
//switch c.Direction {
//case Left:
//v.CCursor -= c.Repeat
//case Up:
//v.RCursor -= c.Repeat
//case Right:
//v.CCursor += c.Repeat
//case Down:
//v.RCursor += c.Repeat
//}

//v.XCursor, v.YCursor = v.CCursor, v.RCursor
//log.Debug("View:%v", v)
//ed.PubEvent("updateView", v)

//return nil
//}

func moveCursor(ed *Editor, kp key.KeyPress) (bool, error) {
	v := ed.ActiveView()
	switch kp.Key {
	case key.Left:
		v.CCursor -= 1
	case key.Up:
		v.RCursor -= 1
	case key.Right:
		v.CCursor += 1
	case key.Down:
		v.RCursor += 1
	default:
	}

	v.XCursor, v.YCursor = v.CCursor, v.RCursor
	log.Debug("View:%v", v)
	pubEvent("updateView", v)

	return false, nil
}

package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
)

type Cell struct {
	Char   rune
	Fg, Bg uint16
}

type View struct {
	RCursor, CCursor int //cursor position in the buffer(row,column)
	XCursor, YCursor int //cursor position in the screen
	Contents         [][]rune
}

func (v *View) updateView() {
	// notify all ui
	eachUI(func(ui *UI) {
		w, h := ui.Width, ui.Height
		row := h
		if row > len(v.Contents) {
			row = len(v.Contents)
		}
		vv := View{RCursor: v.RCursor, CCursor: v.CCursor, Contents: make([][]rune, row)}

		xOffset, yOffset := 0, 0
		if h < v.RCursor+1 {
			yOffset = v.RCursor + 1 - h
		}
		vv.YCursor = v.RCursor - yOffset
		if w < v.CCursor+1 {
			xOffset = v.CCursor + 1 - w
			// '\t' will occupy multi cell
		}
		vv.XCursor = v.CCursor - xOffset
		for i := 0; i < row; i++ {
			// wrap off
			line := v.Contents[i+yOffset]
			log.Debug("i:%d, xOffset:%d, yOffset:%d, line:%v", i, xOffset, yOffset, line)
			if xOffset <= len(line) {
				vv.Contents[i] = line[xOffset:]
			} else {
				vv.Contents[i] = []rune{}
			}
		}

		log.Debug("update view:%v", vv)
		pubEvent("updateView", vv)
	})
}

func moveCursor(ed *Editor, kp key.KeyPress) (bool, error) {
	b := ed.ActiveBuf()
	switch kp.Key {
	case key.Left:
		if b.CCursor > 0 {
			b.CCursor -= 1
		} else if b.RCursor > 0 {
			b.RCursor -= 1
			b.CCursor = b.CurrLineChars() - 1
			if b.CCursor < 0 {
				b.CCursor = 0
			}
		}
	case key.Up:
		if b.RCursor > 0 {
			b.RCursor -= 1
		}
		if b.CCursor >= b.CurrLineChars() {
			b.CCursor = b.CurrLineChars() - 1
			if b.CCursor < 0 {
				b.CCursor = 0
			}
		}
	case key.Right:
		if b.CCursor+1 < b.CurrLineChars() {
			b.CCursor += 1
		} else if b.RCursor+1 < b.Lines() {
			b.RCursor += 1
			b.CCursor = 0
		}
	case key.Down:
		if b.RCursor+1 < b.Lines() {
			b.RCursor += 1
		}
		if b.CCursor >= b.CurrLineChars() {
			b.CCursor = b.CurrLineChars() - 1
			if b.CCursor < 0 {
				b.CCursor = 0
			}
		}
	default:
	}

	b.updateView()

	return false, nil
}

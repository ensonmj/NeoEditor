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
	buff             *Buffer
	RCursor, CCursor int //cursor position in the buffer(unit: char)
	XCursor, YCursor int //cursor position in the screen(unit: cell)
	Contents         [][]rune
}

func (v *View) updateView() {
	ts := v.buff.getConfValueInt("tabstop")

	// notify all ui
	eachUI(func(ui *UI) {
		w, h := ui.Width, ui.Height
		row := h
		if row > len(v.Contents) {
			row = len(v.Contents)
		}
		vv := View{RCursor: v.RCursor, CCursor: v.CCursor, Contents: make([][]rune, row)}

		// calc cursor position in screen
		xOffset, yOffset := 0, 0
		if h < v.RCursor+1 {
			yOffset = v.RCursor + 1 - h
		}
		vv.YCursor = v.RCursor - yOffset

		// <TAB> will occupy multi cell
		// cursor always on the first char for <TAB>
		numTab, _ := expandTab(v.Contents[v.RCursor][:v.CCursor], ts)
		log.Debug("number of tab before cursor in current line:%d", numTab)
		if w < v.CCursor+1+numTab*(ts-1) {
			xOffset = v.CCursor + 1 - w + numTab*(ts-1)
		}
		vv.XCursor = v.CCursor + numTab*(ts-1) - xOffset

		// fill chars for display
		for i := 0; i < row; i++ {
			// wrap off
			_, line := expandTab(v.Contents[i+yOffset], ts)
			if xOffset <= len(line) {
				vv.Contents[i] = line[xOffset:]
			} else {
				vv.Contents[i] = []rune{}
			}
			log.Debug("i:%d, xOffset:%d, yOffset:%d, orig line:%v, new line:%v",
				i, xOffset, yOffset, line, vv.Contents[i])
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

// returen number of <TAB> in line and expanded line
func expandTab(line []rune, tabstop int) (int, []rune) {
	num := 0
	echoLine := make([]rune, 0, len(line))

	for _, r := range line {
		if r == '\t' {
			num++
			for i := 0; i < tabstop; i++ {
				echoLine = append(echoLine, ' ')
			}
		} else {
			echoLine = append(echoLine, r)
		}
	}

	return num, echoLine
}

func lenOfLine(line []rune, tabstop int) int {
	_, expandedLine := expandTab(line, tabstop)
	return len(expandedLine)
}

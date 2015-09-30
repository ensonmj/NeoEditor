package main

import (
	ned "github.com/ensonmj/NeoEditor/backend"
	"github.com/nsf/termbox-go"
)

func drawSpliter(l ned.Line) {
	fg, bg := termbox.ColorWhite, termbox.ColorBlack
	x, y := l.X, l.Y
	switch l.Direction {
	case ned.Horizontal:
		for i := 0; i < l.Length; i++ {
			x = x + i
			termbox.SetCell(x, y, '-', fg, bg)
		}
	case ned.Vertical:
		for i := 0; i < l.Length; i++ {
			y = y + i
			termbox.SetCell(x, y, '|', fg, bg)
		}
	}
}

func drawView(v ned.View) {
	fg, bg := termbox.ColorWhite, termbox.ColorBlack
	termbox.Clear(fg, bg)

	text := v.Contents
	cursorOnText := false
	for row, line := range text {
		for col, r := range line {
			fg, bg := termbox.ColorWhite, termbox.ColorBlack

			// cursor display
			if col == v.XCursor && row == v.YCursor {
				// block style cursor
				fg = fg | termbox.AttrReverse
				cursorOnText = true
			}

			termbox.SetCell(col, row, r, fg, bg)
		}
	}

	if !cursorOnText {
		// block style cursor
		fg = fg | termbox.AttrReverse
		termbox.SetCell(v.XCursor, v.YCursor, ' ', fg, bg)
	}
	termbox.Flush()
}

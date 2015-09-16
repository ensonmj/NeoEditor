package iface

type View struct {
	xOffset          int //must be 0 when wrap is on
	yOffset          int //the line displayed in the top of window
	rCursor, cCursor int //cursor position in the buffer(row,column)
	xCursor, yCursor int //cursor position in the screen
	//xUpdate, yUpdate int //start position for redraw
	Contents [][]rune
}

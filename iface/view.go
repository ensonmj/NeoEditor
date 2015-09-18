package iface

type View struct {
	XOffset          int //must be 0 when wrap is on
	YOffset          int //the line displayed in the top of window
	RCursor, CCursor int //cursor position in the buffer(row,column)
	XCursor, YCursor int //cursor position in the screen
	//xUpdate, yUpdate int //start position for redraw
	Contents [][]rune
}

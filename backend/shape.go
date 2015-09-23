package neoeditor

type Direction int

const (
	NoDirect Direction = iota
	Horizontal
	Vertical
	Left
	Up
	Right
	Down
)

type Point struct {
	X, Y int
}

type Line struct {
	Point // the left or up point of line
	Direction
	Length int
}

type Rect struct {
	Point         // the left-up point of rect
	Width, Height int
}

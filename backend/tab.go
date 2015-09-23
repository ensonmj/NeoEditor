package neoeditor

type Tab struct {
	Id        int
	Wnds      []*Window
	ActiveWnd int
}

func NewTab(seq int) *Tab {
	return &Tab{Id: seq, Wnds: make([]*Window, 1)}
}

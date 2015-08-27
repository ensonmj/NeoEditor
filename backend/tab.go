package neoeditor

type Tab struct {
	id        int
	wnds      []*Window
	activeWnd int
}

func NewTab(seq int) *Tab {
	return &Tab{id: seq, wnds: make([]*Window, 1)}
}

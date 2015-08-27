package neoeditor

type Editor struct {
	tabs      []*Tab
	activeTab int
	bufs      []*Buffer
	activeBuf int
	plugins   []Plugin
}

type Plugin interface {
}

func NewEditor() *Editor {
	return &Editor{tabs: make([]*Tab, 1)}
}

func (ed *Editor) handleInput(ch rune) {

}

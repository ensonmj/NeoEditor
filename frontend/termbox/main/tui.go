package main

import (
	"github.com/ensonmj/NeoEditor/lib/plugin"
	"github.com/nsf/termbox-go"
)

type TUI struct {
	plugin.DummyPlugin
	doRender chan bool
}

func (ui *TUI) Init(name, guid string) {
	ui.Name, ui.Guid = name, guid
	ui.doRender = make(chan bool)
}

func (ui *TUI) Handle(pi *plugin.PluginInput) (*plugin.PluginOutput, error) {
	render := func(text [][]rune) {
		fg, bg := termbox.ColorDefault, termbox.ColorDefault
		termbox.Clear(fg, bg)

		for row, line := range text {
			for col, r := range line {
				termbox.SetCell(col, row, r, termbox.ColorWhite, termbox.ColorDefault)
			}
		}
		termbox.Flush()
	}

	text := pi.Text
	go render(text)

	return nil, nil
}

func (ui *TUI) Render() {
	ui.doRender <- true
}

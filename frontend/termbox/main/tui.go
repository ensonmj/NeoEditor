package main

import (
	ned "github.com/ensonmj/NeoEditor/backend"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/nsf/termbox-go"
)

type TUI struct {
	doRender chan bool
	ed       *ned.Editor
}

func (ui *TUI) Init() {
	log.Debug("TUI Init")
	ui.doRender = make(chan bool)
	ui.ed = ned.Ned
	ui.ed.RegisterListener("bufferChanged", ui)
}

func (ui *TUI) HandleInput(kp key.KeyPress) {
	ui.ed.HandleInput(kp)
}

func (ui *TUI) Render() {
	ui.doRender <- true
}

func (ui *TUI) OnEvent(args ...interface{}) {
	go func(args ...interface{}) {
		log.Debug("TUI get event:%#v", args)
		if args == nil {
			return
		}
		if text, ok := args[0].([][]rune); ok {
			fg, bg := termbox.ColorDefault, termbox.ColorDefault
			termbox.Clear(fg, bg)
			for row, line := range text {
				for col, r := range line {
					termbox.SetCell(col, row, r, termbox.ColorWhite,
						termbox.ColorDefault)
				}
			}
			termbox.Flush()
		}
	}(args...)
}

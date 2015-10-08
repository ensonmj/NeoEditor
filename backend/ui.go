package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
)

type UI struct {
	Id            int
	Width, Height int
}

var uiManager struct {
	uis      []*UI
	activeUI int
}

func registerUI(ui *UI) {
	// calc ID
	ui.Id = cap(uiManager.uis)
	uiManager.activeUI = len(uiManager.uis)
	uiManager.uis = append(uiManager.uis, ui)
}

func activeUI() *UI {
	return uiManager.uis[uiManager.activeUI]
}

func eachUI(f func(*UI)) {
	for _, ui := range uiManager.uis {
		f(ui)
	}
}

// Command
type CmdRegisterUI struct{}

func (c CmdRegisterUI) Run(ed *Editor, args codec.RawMessage) (bool, error) {
	log.Debug("run command [RegisterUI]")
	return false, nil
}

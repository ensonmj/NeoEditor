package main

import (
	"flag"
	"time"

	ned "github.com/ensonmj/NeoEditor/backend"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/ensonmj/NeoEditor/lib/plugin"
	"github.com/nsf/termbox-go"
)

var shutdown chan bool
var keyCh []rune
var (
	lut = map[termbox.Key]ned.KeyPress{
		// Omission of these are intentional due to map collisions
		//		termbox.KeyCtrlTilde:      keys.KeyPress{Ctrl: true, Key: '~'},
		//		termbox.KeyCtrlBackslash:  keys.KeyPress{Ctrl: true, Key: '\\'},
		//		termbox.KeyCtrlSlash:      keys.KeyPress{Ctrl: true, Key: '/'},
		//		termbox.KeyCtrlUnderscore: keys.KeyPress{Ctrl: true, Key: '_'},
		//		termbox.KeyCtrlLsqBracket: keys.KeyPress{Ctrl: true, Key: '{'},
		//		termbox.KeyCtrlRsqBracket: keys.KeyPress{Ctrl: true, Key: '}'},
		// termbox.KeyCtrl3:
		// termbox.KeyCtrl8
		//		termbox.KeyCtrl2:      keys.KeyPress{Ctrl: true, Key: '2'},
		termbox.KeyCtrlSpace: {Ctrl: true, Key: ' '},
		termbox.KeyCtrlA:     {Ctrl: true, Key: 'a'},
		termbox.KeyCtrlB:     {Ctrl: true, Key: 'b'},
		termbox.KeyCtrlC:     {Ctrl: true, Key: 'c'},
		termbox.KeyCtrlD:     {Ctrl: true, Key: 'd'},
		termbox.KeyCtrlE:     {Ctrl: true, Key: 'e'},
		termbox.KeyCtrlF:     {Ctrl: true, Key: 'f'},
		termbox.KeyCtrlG:     {Ctrl: true, Key: 'g'},
		termbox.KeyCtrlH:     {Ctrl: true, Key: 'h'},
		termbox.KeyCtrlJ:     {Ctrl: true, Key: 'j'},
		termbox.KeyCtrlK:     {Ctrl: true, Key: 'k'},
		termbox.KeyCtrlL:     {Ctrl: true, Key: 'l'},
		termbox.KeyCtrlN:     {Ctrl: true, Key: 'n'},
		termbox.KeyCtrlO:     {Ctrl: true, Key: 'o'},
		termbox.KeyCtrlP:     {Ctrl: true, Key: 'p'},
		termbox.KeyCtrlQ:     {Ctrl: true, Key: 'q'},
		termbox.KeyCtrlR:     {Ctrl: true, Key: 'r'},
		termbox.KeyCtrlS:     {Ctrl: true, Key: 's'},
		termbox.KeyCtrlT:     {Ctrl: true, Key: 't'},
		termbox.KeyCtrlU:     {Ctrl: true, Key: 'u'},
		termbox.KeyCtrlV:     {Ctrl: true, Key: 'v'},
		termbox.KeyCtrlW:     {Ctrl: true, Key: 'w'},
		termbox.KeyCtrlX:     {Ctrl: true, Key: 'x'},
		termbox.KeyCtrlY:     {Ctrl: true, Key: 'y'},
		termbox.KeyCtrlZ:     {Ctrl: true, Key: 'z'},
		termbox.KeyCtrl4:     {Ctrl: true, Key: '4'},
		termbox.KeyCtrl5:     {Ctrl: true, Key: '5'},
		termbox.KeyCtrl6:     {Ctrl: true, Key: '6'},
		termbox.KeyCtrl7:     {Ctrl: true, Key: '7'},
		// termbox.KeyEnter:      {Key: keys.Enter},
		// termbox.KeySpace:      {Key: ' '},
		// termbox.KeyBackspace2: {Key: keys.Backspace},
		// termbox.KeyArrowUp:    {Key: keys.Up},
		// termbox.KeyArrowDown:  {Key: keys.Down},
		// termbox.KeyArrowLeft:  {Key: keys.Left},
		// termbox.KeyArrowRight: {Key: keys.Right},
		// termbox.KeyDelete:     {Key: keys.Delete},
		// termbox.KeyEsc:        {Key: keys.Escape},
		// termbox.KeyPgup:       {Key: keys.PageUp},
		// termbox.KeyPgdn:       {Key: keys.PageDown},
		// termbox.KeyF1:         {Key: keys.F1},
		// termbox.KeyF2:         {Key: keys.F2},
		// termbox.KeyF3:         {Key: keys.F3},
		// termbox.KeyF4:         {Key: keys.F4},
		// termbox.KeyF5:         {Key: keys.F5},
		// termbox.KeyF6:         {Key: keys.F6},
		// termbox.KeyF7:         {Key: keys.F7},
		// termbox.KeyF8:         {Key: keys.F8},
		// termbox.KeyF9:         {Key: keys.F9},
		// termbox.KeyF10:        {Key: keys.F10},
		// termbox.KeyF11:        {Key: keys.F11},
		// termbox.KeyF12:        {Key: keys.F12},
		termbox.KeyTab: {Key: '\t'},
	}
)

func main() {
	flag.Parse()

	log.AddFilter("termbox", log.DEBUG, log.NewFileLogWriter("./ned.log"))
	log.Debug("NeoEditor started")
	defer log.Debug("NeoEditor quit")

	shutdown = make(chan bool, 1)
	keyCh = make([]rune, 0)

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	evchan := make(chan termbox.Event, 32)
	go func() {
		for {
			evchan <- termbox.PollEvent()
		}
	}()

	tui := &TUI{}
	tickChan := time.NewTicker(1 * time.Millisecond).C
	ed := ned.NewEditor()
	for {
		select {
		case ev := <-evchan:
			switch ev.Type {
			case termbox.EventError:
				return
			case termbox.EventKey:
				handleInput(ed, ev)
			}
		case <-shutdown:
			return
		case <-tickChan:
			pi := &plugin.PluginInput{}
			tui.Handle(pi)
		}
	}
}

func handleInput(ed *ned.Editor, ev termbox.Event) {
	if ev.Key == termbox.KeyCtrlQ {
		shutdown <- true
		return
	}

	var kp ned.KeyPress
	if ev.Ch != 0 {
		kp.Key = ned.Key(ev.Ch)
		kp.Text = string(ev.Ch)
	} else {
		var ok bool
		// kp is “zero value" if not found in map
		kp, ok = lut[ev.Key]
		log.Debug("key press:%v, ok:%v", kp, ok)
		if !ok {
			kp.Key = ned.Key(ev.Ch)
		}
		kp.Text = string(ev.Ch)
	}

	log.Debug("key press:%v", kp)
	ed.HandleInput(kp)

	keyCh = append(keyCh, ev.Ch)
}

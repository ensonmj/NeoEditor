package main

import (
	"flag"

	"github.com/nsf/termbox-go"
)

var shutdown chan bool
var keyCh []rune

func main() {
	flag.Parse()

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

	for {
		select {
		case ev := <-evchan:
			switch ev.Type {
			case termbox.EventError:
				return
			case termbox.EventKey:
				handleInput(ev)
			}
		case <-shutdown:
			return
		}
	}
}

func redraw() {
	const colordef = termbox.ColorDefault
	termbox.Clear(colordef, colordef)

	for i, r := range keyCh {
		termbox.SetCell(i, 0, r, termbox.ColorWhite, colordef)
	}
	termbox.Flush()
}

func handleInput(ev termbox.Event) {
	if ev.Key == termbox.KeyCtrlQ {
		shutdown <- true
		return
	}

	keyCh = append(keyCh, ev.Ch)
	redraw()
}

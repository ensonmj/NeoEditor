package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/ensonmj/NeoEditor/backend"
	"github.com/ensonmj/NeoEditor/iface"
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/nsf/termbox-go"
	zmq "github.com/pebbe/zmq4"
)

type UI struct {
	width, height int
}

const (
	chanBufLen = 16
)

var (
	shutdown chan bool
	cmdChan  chan string
	ui       UI
	lut      = map[termbox.Key]key.KeyPress{
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
		termbox.KeyCtrlSpace:  {Ctrl: true, Key: ' '},
		termbox.KeyCtrlA:      {Ctrl: true, Key: 'a'},
		termbox.KeyCtrlB:      {Ctrl: true, Key: 'b'},
		termbox.KeyCtrlC:      {Ctrl: true, Key: 'c'},
		termbox.KeyCtrlD:      {Ctrl: true, Key: 'd'},
		termbox.KeyCtrlE:      {Ctrl: true, Key: 'e'},
		termbox.KeyCtrlF:      {Ctrl: true, Key: 'f'},
		termbox.KeyCtrlG:      {Ctrl: true, Key: 'g'},
		termbox.KeyCtrlH:      {Ctrl: true, Key: 'h'},
		termbox.KeyCtrlJ:      {Ctrl: true, Key: 'j'},
		termbox.KeyCtrlK:      {Ctrl: true, Key: 'k'},
		termbox.KeyCtrlL:      {Ctrl: true, Key: 'l'},
		termbox.KeyCtrlN:      {Ctrl: true, Key: 'n'},
		termbox.KeyCtrlO:      {Ctrl: true, Key: 'o'},
		termbox.KeyCtrlP:      {Ctrl: true, Key: 'p'},
		termbox.KeyCtrlQ:      {Ctrl: true, Key: 'q'},
		termbox.KeyCtrlR:      {Ctrl: true, Key: 'r'},
		termbox.KeyCtrlS:      {Ctrl: true, Key: 's'},
		termbox.KeyCtrlT:      {Ctrl: true, Key: 't'},
		termbox.KeyCtrlU:      {Ctrl: true, Key: 'u'},
		termbox.KeyCtrlV:      {Ctrl: true, Key: 'v'},
		termbox.KeyCtrlW:      {Ctrl: true, Key: 'w'},
		termbox.KeyCtrlX:      {Ctrl: true, Key: 'x'},
		termbox.KeyCtrlY:      {Ctrl: true, Key: 'y'},
		termbox.KeyCtrlZ:      {Ctrl: true, Key: 'z'},
		termbox.KeyCtrl4:      {Ctrl: true, Key: '4'},
		termbox.KeyCtrl5:      {Ctrl: true, Key: '5'},
		termbox.KeyCtrl6:      {Ctrl: true, Key: '6'},
		termbox.KeyCtrl7:      {Ctrl: true, Key: '7'},
		termbox.KeyEnter:      {Key: key.Enter},
		termbox.KeySpace:      {Key: ' '},
		termbox.KeyBackspace2: {Key: key.Backspace},
		termbox.KeyArrowUp:    {Key: key.Up},
		termbox.KeyArrowDown:  {Key: key.Down},
		termbox.KeyArrowLeft:  {Key: key.Left},
		termbox.KeyArrowRight: {Key: key.Right},
		termbox.KeyDelete:     {Key: key.Delete},
		termbox.KeyEsc:        {Key: key.Escape},
		termbox.KeyPgup:       {Key: key.PageUp},
		termbox.KeyPgdn:       {Key: key.PageDown},
		termbox.KeyF1:         {Key: key.F1},
		termbox.KeyF2:         {Key: key.F2},
		termbox.KeyF3:         {Key: key.F3},
		termbox.KeyF4:         {Key: key.F4},
		termbox.KeyF5:         {Key: key.F5},
		termbox.KeyF6:         {Key: key.F6},
		termbox.KeyF7:         {Key: key.F7},
		termbox.KeyF8:         {Key: key.F8},
		termbox.KeyF9:         {Key: key.F9},
		termbox.KeyF10:        {Key: key.F10},
		termbox.KeyF11:        {Key: key.F11},
		termbox.KeyF12:        {Key: key.F12},
		termbox.KeyTab:        {Key: '\t'},
	}
)

// Command line flags
var (
	showDebug = flag.Bool("debug", false, "Display debug log")
)

func main() {
	//log.AddFilter("console", log.DEBUG, log.NewConsoleLogWriter())
	log.Debug("NeoEditor started")
	defer log.Close()

	// For profile
	go func() {
		http.ListenAndServe("127.0.0.1:5199", nil)
	}()

	defer func() {
		termbox.Close()
		if err := recover(); err != nil {
			log.Critical(err)
			log.Close()
			panic(err)
		}
	}()

	flag.Parse()

	if _, err := neoeditor.NewEditor(); err != nil {
		log.Critical("create editor error:%s", err)
		panic(err)
	}

	shutdown = make(chan bool, 1)
	cmdChan = make(chan string, chanBufLen)

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	ui.width, ui.height = termbox.Size()
	evchan := make(chan termbox.Event, chanBufLen)
	go func() {
		for {
			evchan <- termbox.PollEvent()
		}
	}()

	req, _ := zmq.NewSocket(zmq.PUSH)
	req.Connect("tcp://localhost:5198")

	sub, _ := zmq.NewSocket(zmq.SUB)
	//sub.Connect("tcp://localhost:5199")
	sub.Connect("inproc://notification")
	sub.SetSubscribe("updateView")
	//sub.SetSubscribe("")

	// Assuming that all extra arguments are files
	if files := flag.Args(); len(files) > 0 {
		for _, file := range files {
			openFile(file)
		}
	}

	// Receive notification
	go func() {
		log.Debug("start receiving notification")
		for {
			topic, _ := sub.Recv(0)
			msg, _ := sub.Recv(0)
			log.Debug("subscriber got msg:%s%s", topic, msg)
			var v iface.View
			if err := codec.Deserialize([]byte(msg), &v); err != nil {
				log.Critical(err)
				continue
			}
			updateView(v)
		}
	}()

	//tickChan := time.NewTicker(1 * time.Millisecond).C
	for {
		select {
		case ev := <-evchan:
			switch ev.Type {
			case termbox.EventKey:
				handleInput(req, ev)
			case termbox.EventResize:
				handleResize(ev.Height, ev.Width)
			case termbox.EventError:
				log.Critical("key event error:%v", ev)
				return
			}
		case cmd := <-cmdChan:
			req.Send(cmd, zmq.DONTWAIT)
		case <-shutdown:
			log.Debug("NeoEditor quit")
			return
			//case <-tickChan:
		}
	}
}

func handleInput(req *zmq.Socket, ev termbox.Event) {
	if ev.Key == termbox.KeyCtrlQ {
		shutdown <- true
		return
	}

	var kp key.KeyPress
	if ev.Ch != 0 {
		kp.Key = key.Key(ev.Ch)
		kp.Text = string(ev.Ch)
	} else {
		var ok bool
		// kp is “zero value" if not found in map
		kp, ok = lut[ev.Key]
		log.Debug("key press:%v, ok:%v", kp, ok)
		if !ok {
			kp.Key = key.Key(ev.Ch)
		}
		kp.Text = string(kp.Key)
	}

	log.Debug("key press:%v", kp)
	sendCommand(kp)
}

func handleResize(width, height int) {
	ui.width, ui.height = width, height
}

func updateView(v iface.View) {
	log.Debug("View:%v", v)
	fg, bg := termbox.ColorWhite, termbox.ColorBlack
	termbox.Clear(fg, bg)

	text := v.Contents
	x, y := 0, 0
	cursorOnText := false
	for _, line := range text {
		fg, bg := termbox.ColorWhite, termbox.ColorBlack
		for col, r := range line {
			if col < ui.width {
				x = col
			} else {
				// wrap, line may have many screen lines
				x = col - ui.width
				y++
			}
			if x == v.XCursor && y == v.YCursor {
				// block style cursor
				fg = fg | termbox.AttrReverse
				cursorOnText = true
			}
			termbox.SetCell(x, y, r, fg, bg)
		}
		y++
	}

	if !cursorOnText {
		// block style cursor
		fg = fg | termbox.AttrReverse
		termbox.SetCell(v.XCursor, v.YCursor, ' ', fg, bg)
	}
	termbox.Flush()
}

// Command
func sendCommand(v interface{}) {
	cmd := codec.Envelope{Method: "KeyPress", Arguments: v}
	msg, _ := codec.Serialize(cmd)
	log.Debug("send cmd:%s", msg)
	cmdChan <- string(msg)
}

func openFile(fPath string) {
	sendCommand(map[string]string{"fPath": fPath})
}

package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"

	"github.com/ensonmj/NeoEditor/backend"
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
	"github.com/nsf/termbox-go"
	zmq "github.com/pebbe/zmq4"
)

const (
	chanBufLen = 16
)

var shutdown chan bool
var cmdChan chan string
var (
	lut = map[termbox.Key]key.KeyPress{
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

	evchan := make(chan termbox.Event, chanBufLen)
	go func() {
		for {
			evchan <- termbox.PollEvent()
		}
	}()

	req, _ := zmq.NewSocket(zmq.PUSH)
	req.Connect("tcp://localhost:5198")

	sub, _ := zmq.NewSocket(zmq.SUB)
	sub.Connect("tcp://localhost:5199")
	//sub.SetSubscribe("updateView ")
	sub.SetSubscribe("1 ")

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
		}
	}()

	//tickChan := time.NewTicker(1 * time.Millisecond).C
	for {
		select {
		case ev := <-evchan:
			switch ev.Type {
			case termbox.EventError:
				log.Critical("key event error:%v", ev)
				return
			case termbox.EventKey:
				handleInput(req, ev)
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
		kp.Text = string(ev.Ch)
	}

	log.Debug("key press:%v", kp)
	sendCommand(kp)
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

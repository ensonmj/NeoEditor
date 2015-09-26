package neoeditor

import (
	//"os"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
)

type Commander interface {
	Run(ed *Editor, args codec.RawMessage)
}

type CommandManager map[string]Commander

func (cm CommandManager) registerCommands() {
	cm["KeyPress"] = CmdKeyPress{}
}

func (cm CommandManager) dispatchCommand(ed *Editor, cmd string) {
	log.Debug("receive command:%s", cmd)
	var args codec.RawMessage
	env := codec.Envelope{
		Arguments: &args,
	}
	if err := codec.Deserialize([]byte(cmd), &env); err != nil {
		log.Critical(err)
		return
	}
	log.Debug("parse command:{%s, %v}", env.Method, args)

	if cmd, ok := cm[env.Method]; ok {
		log.Debug("receive command [%s]", env.Method)
		cmd.Run(ed, args)
	} else {
		log.Warn("receive unsupported command [%s]", env.Method)
	}
}

// Commands
type CmdKeyPress struct {
}

func (c CmdKeyPress) Run(ed *Editor, args codec.RawMessage) {
	log.Debug("run command [KeyPress]")
	var kp key.KeyPress
	if err := codec.Deserialize(args, &kp); err != nil {
		log.Critical(err)
		return
	}
	log.Debug("parse command [KeyPress] arguments:%v", kp)
	if kp.Ctrl && kp.Key == 'q' {
		close(ed.done)
		return
	}
	if kp.Ctrl && kp.Key == 's' {
		log.Debug("save buffer:%v", ed.bufs[ed.activeBuf])
		ed.bufs[ed.activeBuf].Save()
		return
	}

	// parse keypress
	ed.ResolvMode(kp)
}

//func (ed *Editor) OpenFiles(fPaths []string) {
//n := len(ed.bufs)
//for _, fPath := range fPaths {
//buf, err := NewBuffer(fPath, os.O_RDWR|os.O_CREATE, 0644)
//if err != nil {
//log.Warn("open file[%s] err:%s", fPath, err)
//continue
//}
//ed.bufs = append(ed.bufs, buf)
//}
//ed.activeBuf = n

//v := ed.bufs[ed.activeBuf].View
//log.Debug("View:%v", v)
//ed.PubEvent("updateView", v)
//}

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

var cmdManager = map[string]Commander{}

func registerCommands() {
	cmdManager["KeyPress"] = CmdKeyPress{}
}

func dispatchCommand(ed *Editor, cmd string) {
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

	if cmd, ok := cmdManager[env.Method]; ok {
		log.Debug("receive command [%s]", env.Method)
		cmd.Run(ed, args)
	} else {
		log.Warn("receive unsupported command [%s]", env.Method)
	}
}

// Commands
type CmdKeyPress struct{}

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
	runModeAction(ed, kp)
}

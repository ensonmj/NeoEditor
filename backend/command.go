package neoeditor

import (
	"errors"
	"fmt"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/key"
	"github.com/ensonmj/NeoEditor/lib/log"
)

type Commander interface {
	// return true if program will exit
	Run(ed *Editor, args codec.RawMessage) (bool, error)
}

var cmdManager = map[string]Commander{}

func registerCommands() {
	cmdManager["KeyPress"] = CmdKeyPress{}
}

// return true if program will exit
func dispatchCommand(ed *Editor, cmd string) (bool, error) {
	log.Debug("receive command:%s", cmd)
	var args codec.RawMessage
	env := codec.Envelope{
		Arguments: &args,
	}
	if err := codec.Deserialize([]byte(cmd), &env); err != nil {
		log.Critical(err)
		return false, err
	}
	log.Debug("parse command:{%s, %v}", env.Method, args)

	if cmd, ok := cmdManager[env.Method]; ok {
		log.Debug("receive command [%s]", env.Method)
		return cmd.Run(ed, args)
	} else {
		str := fmt.Sprintf("receive unsupported command [%s]", env.Method)
		log.Warn(str)
		errors.New(str)
	}

	return false, nil
}

// Commands
type CmdKeyPress struct{}

func (c CmdKeyPress) Run(ed *Editor, args codec.RawMessage) (bool, error) {
	log.Debug("run command [KeyPress]")
	var kp key.KeyPress
	if err := codec.Deserialize(args, &kp); err != nil {
		log.Critical(err)
		return false, nil
	}
	log.Debug("parse command [KeyPress] arguments:%v", kp)
	if kp.Ctrl && kp.Key == 'q' {
		close(ed.done)
		return true, nil
	}
	if kp.Ctrl && kp.Key == 's' {
		log.Debug("save buffer:%v", ed.bufs[ed.activeBuf])
		ed.bufs[ed.activeBuf].Save()
		return false, nil
	}

	// parse keypress
	return runModeAction(ed, kp)
}

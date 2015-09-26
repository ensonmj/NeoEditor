package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
)

var events chan codec.Envelope

func initEvent() {
	events = make(chan codec.Envelope, chanBufLen)
}

func pollEvent() chan codec.Envelope {
	return events
}

func pubEvent(topic string, arg interface{}) {
	env := codec.Envelope{Method: topic, Arguments: arg}
	log.Debug("publish event:%v", env)
	events <- env
}

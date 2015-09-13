package neoeditor

import (
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
)

type Event struct {
	topic string
	codec.Envelope
}

func (ed *Editor) PubEvent(topic string, arg interface{}) {
	env := codec.Envelope{Method: topic, Arguments: arg}
	log.Debug("publish event:%v", env)
	ed.events <- env
}

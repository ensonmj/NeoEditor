package events

type BufferChanged struct {
	listeners []Listener
}

func (bc *BufferChanged) AddListener(l Listener) {
	bc.listeners = append(bc.listeners, l)
}

func (bc *BufferChanged) Notify(args ...interface{}) {
	for _, l := range bc.listeners {
		l.OnEvent(args...)
	}
}

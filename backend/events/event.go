package events

type Listener interface {
	OnEvent(args ...interface{})
}

type Event interface {
	AddListener(l Listener)
	Notify(args ...interface{})
}

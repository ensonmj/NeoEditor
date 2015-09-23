package neoeditor

type Edit struct {
	Action     string //"insert","replace","delete"
	Row, Col   int
	Data       []rune
	Prev, Next *Edit
}

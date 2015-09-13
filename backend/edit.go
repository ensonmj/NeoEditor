package neoeditor

type Edit struct {
	action string //"insert","replace","delete"
	line   int
	data   []rune
}

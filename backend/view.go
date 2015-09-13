package neoeditor

type View struct {
	lOffset int  // line offset
	cOffset int  // character offset, may equal several vOffset, e.g. '\t'
	vOffset int  // visual offset
	num     bool // display linenumber
}

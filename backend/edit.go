package neoeditor

import "github.com/ensonmj/NeoEditor/lib/log"

type EdType int

const (
	ETInsert EdType = iota
	ETReplace
	ETDelete
)

type Edit struct {
	Action         EdType
	Row, Col       int
	Text, OrigText []rune
	Prev, Next     *Edit
}

func NewInsertEdit() *Edit {
	return &Edit{Action: ETInsert, Row: -1, Col: -1}
}

// orig may include multi line
func NewReplaceEdit(orig []rune) *Edit {
	e := &Edit{Action: ETReplace, Row: -1, Col: -1}
	e.OrigText = append(e.OrigText, orig...)
	return e
}

func NewDeleteEdit() *Edit {
	return &Edit{Action: ETDelete, Row: -1, Col: -1}
}

func (e *Edit) Save(row, col int, data []rune) {
	if row < 0 || col < 0 {
		log.Warn("save edit row:%d col:%d invalid", row, col)
		return
	}
	if e.Row == -1 && e.Col == -1 {
		e.Row = row
		e.Col = col
	} else if e.Row != row {
		log.Warn("save edit row:%d col:%d not equal prev row:%d", row, col, e.Row)
		return
	}
	switch e.Action {
	case ETInsert:
		// make sure 'edit area' is continuous
		if e.Col+len(e.Text) != col && e.Col != col {
			log.Warn("save edit row:%d col:%d not continuous prev col:%d", row, col, e.Col)
			return
		}
		e.Text = append(e.Text, data...)
	case ETReplace:
		e.Text = append(e.Text, data...)
	case ETDelete:
		e.OrigText = append(e.OrigText, data...)
	}
}

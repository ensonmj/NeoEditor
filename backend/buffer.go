package neoeditor

import (
	"errors"
	"os"

	"github.com/ensonmj/NeoEditor/iface"
	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
)

// default size for one line
const chunkSize = 128

// TODO: store content line by line, and support to highlight diff
type Buffer struct {
	iface.View
	scratch bool
	fPath   string
	file    *os.File
	edits   []*Edit
	data    [][]rune
}

func NewBuffer(fPath string, flag int, perm os.FileMode) (*Buffer, error) {
	if fPath == "" {
		log.Debug("create scratch buffer")
		buffer := &Buffer{scratch: true}
		buffer.data = append(buffer.data, make([]rune, 0, chunkSize))
		return buffer, nil
	}

	fd, err := os.OpenFile(fPath, flag, perm)
	if err != nil {
		return nil, err
	}
	log.Debug("create new file:%s", fPath)

	buffer := &Buffer{fPath: fPath, file: fd}
	buffer.data = append(buffer.data, make([]rune, 0, chunkSize))
	return buffer, nil
}

//func (b *Buffer) String() string {
//return b.fPath + ":" + string(b.data)
//}

func (b *Buffer) Contents() [][]rune {
	log.Debug("buffer contents:%#v", b.data)
	return b.data
}

func (b *Buffer) Insert(chars []rune) error {
	row, col := b.RCursor, b.CCursor
	log.Debug("Insert %s in %d,%d", string(chars), row, col)
	lineStart := 0 // start pos of a batch chars which splited by '\n'
	for i, c := range chars {
		if c == '\n' {
			buf := b.data[row]
			nextLine := make([]rune, 0, chunkSize)
			//copy the ramain to next line
			copy(nextLine, buf[col:])
			b.data = append(b.data, nextLine)

			// '\n' not saved in the buffer
			b.data[row], _ = replaceFrom(buf, chars[lineStart:i], col)
			lineStart = i + 1
			row++
			col = 0
		} else if i == len(chars)-1 {
			buf := b.data[row]
			b.data[row], _ = insertIn(buf, chars[lineStart:], col)
			col = col + len(chars[lineStart:])
			log.Debug("buffer content:%v", b.data)
		} else {
			continue
		}
	}

	b.RCursor, b.CCursor = row, col

	// TODO: calc cursor position
	b.XCursor, b.YCursor = b.CCursor, b.RCursor
	b.View.Contents = b.data

	return nil
}

func replaceFrom(orig, chars []rune, index int) ([]rune, error) {
	if index > len(orig) {
		return nil, errors.New("gap not allowed in file")
	}

	req := index + len(chars)
	if req > cap(orig) {
		alloc := (req + chunkSize - 1) & ^(chunkSize - 1)
		n := make([]rune, index, alloc)
		copy(n, orig[0:index]) //chars from index will be overwrite
		orig = n
	}
	copy(orig[index:req], chars)

	orig = orig[:req]

	return orig, nil
}

func insertIn(orig, chars []rune, index int) ([]rune, error) {
	if index > len(orig) {
		return nil, errors.New("gap not allowed in file")
	}

	req := len(orig) + len(chars)
	if req > cap(orig) {
		alloc := (req + chunkSize - 1) & ^(chunkSize - 1)
		n := make([]rune, len(orig), alloc)
		copy(n, orig)
		orig = n
	}

	// append chars into data
	if index == len(orig) {
		copy(orig[len(orig):req], chars)
	} else {
		copy(orig[index+len(chars):cap(orig)], orig[index:len(orig)])
		copy(orig[index:req], chars)
	}
	orig = orig[:req]

	return orig, nil
}

func (b *Buffer) UnInsert(index int, chars []rune) error {
	return nil
}

func (b *Buffer) Surround(start, end int, fChars, bChars []rune) error {
	return nil
}

func (b *Buffer) Close() {
	log.Debug("Close the buffer:%s", b.fPath)
	//b.file.Write([]byte(string(b.data)))
	b.file.Close()
}

// Commands
type CmdNewBuffer struct {
	fPath string
	flag  int
	perm  os.FileMode
}

func (c CmdNewBuffer) Run(ed *Editor, args string) error {
	codec.Deserialize([]byte(args), c)
	buf, err := NewBuffer(c.fPath, c.flag, c.perm)
	if err != nil {
		return err
	}
	ed.bufs = append(ed.bufs, buf)
	ed.activeBuf = len(ed.bufs) - 1

	return nil
}

type CmdInsertRune struct {
	data string
}

func (c CmdInsertRune) Run(ed *Editor) error {
	log.Debug("InsertRune data:%s", c.data)
	//codec.Deserialize([]byte(args), c)
	//log.Debug("after parse:%v", c)
	ed.bufs[ed.activeBuf].Insert([]rune(c.data))

	v := ed.bufs[ed.activeBuf].View
	log.Debug("View:%v", v)
	ed.PubEvent("updateView", v)

	return nil
}

type CursorDirection int

const (
	CLeft CursorDirection = iota
	CUp
	CRight
	CDown
)

type CmdMoveCursor struct {
	Direction CursorDirection
	Repeat    int
}

func (c CmdMoveCursor) Run(ed *Editor) error {
	v := &ed.bufs[ed.activeBuf].View
	switch c.Direction {
	case CLeft:
		v.CCursor -= c.Repeat
	case CUp:
		v.RCursor -= c.Repeat
	case CRight:
		v.CCursor += c.Repeat
	case CDown:
		v.RCursor += c.Repeat
	}

	v.XCursor, v.YCursor = v.CCursor, v.RCursor
	log.Debug("View:%v", v)
	ed.PubEvent("updateView", v)

	return nil
}

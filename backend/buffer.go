package neoeditor

import (
	"os"

	"github.com/ensonmj/NeoEditor/lib/codec"
	"github.com/ensonmj/NeoEditor/lib/log"
)

const chunkSize = 256 * 1024

// TODO: store content line by line, and support to highlight diff
type Buffer struct {
	view    *View
	scratch bool
	fPath   string
	file    *os.File
	edits   []*Edit
	data    []rune
}

func NewBuffer(fPath string, flag int, perm os.FileMode) (*Buffer, error) {
	fd, err := os.OpenFile(fPath, flag, perm)
	if err != nil {
		return nil, err
	}
	log.Debug("create new file:%s", fPath)

	return &Buffer{fPath: fPath, file: fd}, nil
}

func (b *Buffer) String() string {
	return b.fPath + ":" + string(b.data)
}

func (b *Buffer) Contents() [][]rune {
	contents := make([][]rune, 1)
	contents[0] = b.data
	log.Debug("buffer contents:%#v", contents)
	return contents
}

func (b *Buffer) Insert(index int, chars []rune) error {
	log.Debug("Insert in:%d,%s", index, string(chars))
	req := len(chars) + len(b.data)
	if req > cap(b.data) {
		alloc := (req + chunkSize - 1) & ^(chunkSize - 1)
		n := make([]rune, len(b.data), alloc)
		copy(n, b.data)
		b.data = n
	}

	// append chars into data
	if index >= len(b.data) {
		// not allowed gap in file
		copy(b.data[len(b.data):req], chars)
	} else {
		copy(b.data[index+len(chars):cap(b.data)], b.data[index:len(b.data)])
		copy(b.data[index:req], chars)
	}
	b.data = b.data[:req]

	return nil
}

func (b *Buffer) UnInsert(index int, chars []rune) error {
	return nil
}

func (b *Buffer) Surround(start, end int, fChars, bChars []rune) error {
	if err := b.Insert(start, fChars); err != nil {
		return err
	}

	index := end + len(fChars)
	if err := b.Insert(index, bChars); err != nil {
		return err
	}

	return nil
}

func (b *Buffer) Append(chars []rune) error {
	index := len(b.data)
	return b.Insert(index, chars)
}

func (b *Buffer) Close() {
	log.Debug("Close the buffer:%s", b.fPath)
	b.file.Write([]byte(string(b.data)))
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

type CmdAppendRune struct {
	data string
}

func (c CmdAppendRune) Run(ed *Editor, args string) error {
	codec.Deserialize([]byte(args), c)
	ed.bufs[ed.activeBuf].Append([]rune(c.data))

	ed.PubEvent("updateView", ed.bufs[ed.activeBuf].Contents())

	return nil
}

// Events
//type BufferChanged struct {
//listeners []Listener
//}

//func (bc *BufferChanged) AddListener(l Listener) {
//bc.listeners = append(bc.listeners, l)
//}

//func (bc *BufferChanged) Notify(args ...interface{}) {
//for _, l := range bc.listeners {
//l.OnEvent("bufferChanged", args...)
//}
//}

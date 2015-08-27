package neoeditor

import "os"

const chunkSize = 256 * 1024

type Buffer struct {
	scratch bool
	file    *os.File
	data    []rune
}

func NewBuffer(fPath string, flag int, perm os.FileMode) (*Buffer, error) {
	fd, err := os.OpenFile(fPath, flag, perm)
	if err != nil {
		return nil, err
	}
	return &Buffer{file: fd}, nil
}

func (b *Buffer) Insert(index int, chars []rune) error {
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

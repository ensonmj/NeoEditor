package neoeditor

import (
	"bufio"
	"errors"
	"os"

	"github.com/ensonmj/NeoEditor/lib/log"
)

// default size for one line
const chunkSize = 128

// TODO: store content line by line, and support to highlight diff
type Buffer struct {
	View
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

	fi, _ := fd.Stat()
	if fi.Size() == 0 {
		// new or empty file
		buffer.data = append(buffer.data, make([]rune, 0, chunkSize))
	} else {
		// read file contents to data
		scanner := bufio.NewScanner(fd)
		for scanner.Scan() {
			line := scanner.Text()
			buffer.data = append(buffer.data, []rune(line))
		}
	}

	buffer.View.Contents = buffer.data
	buffer.updateView()

	return buffer, nil
}

func (b Buffer) Contents() [][]rune {
	log.Debug("buffer contents:%#v", b.data)
	return b.data
}

func (b Buffer) Lines() int {
	return len(b.data)
}

func (b Buffer) CurrLineChars() int {
	return len(b.data[b.RCursor])
}

func (b *Buffer) Insert(char rune) error {
	row, col := b.RCursor, b.CCursor
	log.Debug("Insert %s in %d,%d", string(char), row, col)
	// '\n' not saved in the buffer
	if char == '\n' {
		buf := b.data[row]

		//copy the ramain to next line
		nextLine := make([]rune, len(buf[col:]), chunkSize)
		copy(nextLine, buf[col:])

		//update current row
		b.data[row] = buf[:col]
		log.Debug("enter split:[%v][%v]", b.data[row], nextLine)

		// insert the nextLine into data
		b.data = append(b.data, nil)
		copy(b.data[row+2:], b.data[row+1:])
		b.data[row+1] = nextLine
		log.Debug("all data:%v", b.data)

		// update row and col
		row++
		col = 0
	} else {
		buf := b.data[row]
		b.data[row], _ = insertIn(buf, []rune{char}, col)
		col++
	}
	b.RCursor, b.CCursor = row, col

	v := b.View
	v.Contents = b.data
	v.updateView()

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

func (b *Buffer) Save() {
	log.Debug("save the buffer:%s", b.fPath)
	b.file.Seek(0, 0)
	for _, line := range b.data {
		b.file.Write([]byte(string(line) + "\n"))
	}
	b.file.Sync()
}

func (b *Buffer) Close() {
	log.Debug("close the buffer:%s", b.fPath)
	b.file.Seek(0, 0)
	for _, line := range b.data {
		b.file.Write([]byte(string(line) + "\n"))
	}
	b.file.Close()
}

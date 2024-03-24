package io

import "fmt"

type OutOfBoundsError struct { Index, Capacity int }
func (o OutOfBoundsError) Error() string {
	return fmt.Sprintf("io/byte_buffer: out of bounds of internal buffer (index: %d, capacity: %d, index range: 0..%d)", o.Index, o.Capacity, o.Capacity - 1)
}

type ByteBuffer struct {
	Buf []byte // internal byte array
	Position int
	initialSize int
	maxWritten int
}

func NewByteBuffer(size int) *ByteBuffer {
	return &ByteBuffer{Buf: make([]byte, size), Position: 0, initialSize: size}
}

func (bb *ByteBuffer) Get(pos int) (byte, error) {
	if err := bb.checkIndex(pos); err != nil {
		return 0, err
	}
	return bb.Buf[pos], nil
}

func (bb *ByteBuffer) Put(pos int, val byte) error {
	if err := bb.checkIndex(pos); err != nil {
		return err
	}
	bb.Buf[pos] = val
	return nil
}

func (bb *ByteBuffer) Write(val byte) error {
	if err := bb.checkIndex(bb.Position); err != nil {
		return err
	}
	bb.Buf[bb.Position] = val
	bb.Position++
	bb.maxWritten++
	return nil
}

func (bb *ByteBuffer) Read() (byte, error) {
	if err := bb.checkIndex(bb.Position); err != nil {
		return 0, err
	}
	val := bb.Buf[bb.Position]
	bb.Position++
	return val, nil
}

func (bb *ByteBuffer) Cap() int {
	return cap(bb.Buf)
}

func (bb *ByteBuffer) Len() int {
	return len(bb.Buf)
}

func (bb *ByteBuffer) Remaining() int {
	return bb.maxWritten - bb.Position
}

func (bb *ByteBuffer) Resize(size int) {
	if size < bb.Position {
		bb.Position = size
	}
	tmp := make([]byte, size)
	copy(tmp, bb.Buf)
	bb.Buf = tmp
}

func (bb *ByteBuffer) Append(newBytes []byte) {
	//if len(newBytes) > bb.Remaining() {
	//	bb.Resize(bb.Cap() - bb.Remaining() + len(newBytes))
	//}
	for _, b := range newBytes {
		bb.Write(b)
	}
}

func (bb *ByteBuffer) Flip() {
	bb.Position = 0
}

func (bb *ByteBuffer) Compact() {
	bb.Buf = bb.Buf[bb.Position:]
	bb.maxWritten -= bb.Position
	bb.Flip()
	bb.Resize(bb.initialSize)
}

func (bb *ByteBuffer) Buffer() []byte {
	return bb.Buf[:bb.maxWritten]
}

func (bb *ByteBuffer) checkIndex(pos int) error {
	if pos < 0 || pos >= bb.Len() {
		return OutOfBoundsError{Capacity: bb.Cap(), Index: pos}
	}
	return nil
}
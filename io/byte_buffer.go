package io

import "fmt"

type OutOfBoundsError struct { Index, Capacity int }
func (o OutOfBoundsError) Error() string {
	return fmt.Sprintf("io/byte_buffer: out of bounds of internal buffer (index: %d, capacity: %d, index range: 0..%d)", o.Index, o.Capacity, o.Capacity - 1)
}

type ByteBuffer struct {
	buf []byte // internal byte array
	Position int
}

func NewByteBuffer(size int) *ByteBuffer {
	return &ByteBuffer{buf: make([]byte, size), Position: 0}
}

func (bb *ByteBuffer) Get(pos int) (byte, error) {
	if err := bb.checkIndex(pos); err != nil {
		return 0, err
	}
	return bb.buf[pos], nil
}

func (bb *ByteBuffer) Put(pos int, val byte) error {
	if err := bb.checkIndex(pos); err != nil {
		return err
	}
	bb.buf[pos] = val
	return nil
}

func (bb *ByteBuffer) Write(val byte) error {
	if err := bb.checkIndex(bb.Position); err != nil {
		return err
	}
	bb.buf[bb.Position] = val
	bb.Position++
	return nil
}

func (bb *ByteBuffer) Cap() int {
	return cap(bb.buf)
}

func (bb *ByteBuffer) Remaining() int {
	return bb.Cap() - bb.Position
}

func (bb *ByteBuffer) Buffer() []byte {
	return bb.buf
}

func (bb *ByteBuffer) Resize(size int) {
	if size < bb.Position {
		bb.Position = size
	}
	tmp := make([]byte, size)
	copy(tmp, bb.buf)
	bb.buf = tmp
}

func (bb *ByteBuffer) checkIndex(pos int) error {
	if pos < 0 || pos >= bb.Cap() {
		return OutOfBoundsError{Capacity: bb.Cap(), Index: pos}
	}
	return nil
}
package io

import (
	"errors"
	"rs-go-server/crypto"
)

var ErrIllegalAccessType = errors.New("io/stream_buffer: illegal access type")

type AccessType int
const (
	BYTE_ACCESS AccessType = iota
	BIT_ACCESS
)

type ValueType int
const (
	STANDARD ValueType = iota
	A
	C
	S
)

type ByteOrder int
const (
	LITTLE ByteOrder = iota
	BIG
	MIDDLE
	INVERSE_MIDDLE
)

var bitmask = [...]int {0, 0x1, 0x3, 0x7, 0xf, 0x1f, 0x3f, 0x7f, 0xff, 0x1ff, 0x3ff, 0x7ff, 0xfff, 0x1fff, 0x3fff, 0x7fff, 0xffff, 0x1ffff, 0x3ffff, 0x7ffff, 0xfffff, 0x1fffff, 0x3fffff, 0x7fffff, 0xffffff, 0x1ffffff, 0x3ffffff, 0x7ffffff, 0xfffffff, 0x1fffffff, 0x3fffffff, 0x7fffffff, -1}

type StreamBuffer struct {
	accessType AccessType
	Buffer *ByteBuffer
	size int
	bitPosition int
	lengthPosition int
}

func NewOutBuffer(size int) *StreamBuffer {
	buffer := NewByteBuffer(size)
	sb := &StreamBuffer{size: size, Buffer: buffer}
	return sb
}

func (sb *StreamBuffer) WriteHeader(cipher *crypto.ISAACCipher, value int) {
	sb.WriteByte(value + int(cipher.Next()), STANDARD)
}

func (sb *StreamBuffer) WriteVariablePacketHeader(cipher *crypto.ISAACCipher, value int) {
	sb.WriteHeader(cipher, value)
	sb.lengthPosition = sb.Buffer.Position
	sb.WriteByte(0, STANDARD)
}

func (sb *StreamBuffer) WriteVariableShortPacketHeader(cipher *crypto.ISAACCipher, value int) {
	sb.WriteHeader(cipher, value)
	sb.lengthPosition = sb.Buffer.Position
	sb.WriteShort(0, STANDARD, BIG)
}

func (sb *StreamBuffer) FinishVariablePacketHeader() {
	sb.Buffer.Put(sb.lengthPosition, byte(sb.Buffer.Position - sb.lengthPosition - 1))
}

func (sb *StreamBuffer) FinishVariableShortPacketHeader() {
	sb.Buffer.Put(sb.lengthPosition, byte(sb.Buffer.Position - sb.lengthPosition -2 ))
}

func (sb *StreamBuffer) WriteBytes(buf *ByteBuffer) {
	for _, b := range buf.Buffer() {
		sb.WriteByte(int(b), STANDARD)
	}
}

func (sb *StreamBuffer) WriteBytesReverse(buf *ByteBuffer) {
	for i := buf.Cap() - 1; i >= 0; i-- {
		sb.WriteByte(int(buf.Buffer()[i]), STANDARD)
	}
}

func (sb *StreamBuffer) WriteBits(amount, value int) error {
	if sb.accessType != BIT_ACCESS {
		return ErrIllegalAccessType
	}

	bytePos := sb.bitPosition >> 3
	bitOffset := 8 - (sb.bitPosition & 7)
	sb.bitPosition += amount

	requiredSpace := bytePos - sb.Buffer.Position + 1
	requiredSpace += (amount + 7) / 8
	if sb.Buffer.Remaining() < requiredSpace {
		sb.Buffer.Resize(sb.Buffer.Cap() + requiredSpace)
	}

	for ; amount > bitOffset; bitOffset = 8 {
		tmp, _ := sb.Buffer.Get(bytePos)
		tmp &= byte(^bitmask[bitOffset])
		tmp |= byte((value >> uint(amount - bitOffset)) & bitmask[bitOffset])
		sb.Buffer.Put(bytePos, tmp)
		bytePos++
		amount -= bitOffset
	}
	if amount == bitOffset {
		tmp, _ := sb.Buffer.Get(bytePos)
		tmp &= byte(^bitmask[bitOffset])
		tmp |= byte(value & bitmask[bitOffset])
		sb.Buffer.Put(bytePos, tmp)
	} else {
		tmp, _ := sb.Buffer.Get(bytePos)
		tmp &= ^byte(bitmask[amount] << uint(bitOffset - amount))
		tmp |= byte((value & bitmask[amount]) << uint(bitOffset - amount))
		sb.Buffer.Put(bytePos, tmp)
	}

	return nil
}

func (sb *StreamBuffer) WriteBit(flag bool) {
	bit := 0
	if flag {
		bit = 1
	}
	sb.WriteBits(1, bit)
}

func (sb *StreamBuffer) WriteByte(value int, valueType ValueType) {
	switch valueType {
	case A:
		value += 128
	case C:
		value = -value
	case S:
		value = 128 - value
	}
	sb.Buffer.Write(byte(value))
}

func (sb *StreamBuffer) WriteShort(value int, valueType ValueType, order ByteOrder) {
	switch order {
	case BIG:
		sb.WriteByte(value >> 8, STANDARD)
		sb.WriteByte(value, valueType)
	case LITTLE:
		sb.WriteByte(value, valueType)
		sb.WriteByte(value >> 8, STANDARD)
	}
}

func (sb *StreamBuffer) WriteInt(value int, valueType ValueType, order ByteOrder) {
	switch order {
	case BIG:
		sb.WriteByte(value >> 24, STANDARD)
		sb.WriteByte(value >> 16, STANDARD)
		sb.WriteByte(value >> 8, STANDARD)
		sb.WriteByte(value, valueType)
	case MIDDLE:
		sb.WriteByte(value >> 8, STANDARD)
		sb.WriteByte(value, valueType)
		sb.WriteByte(value >> 24, STANDARD)
		sb.WriteByte(value >> 16, STANDARD)
	case INVERSE_MIDDLE:
		sb.WriteByte(value >> 16, STANDARD)
		sb.WriteByte(value >> 24, STANDARD)
		sb.WriteByte(value, valueType)
		sb.WriteByte(value >> 8, STANDARD)
	case LITTLE:
		sb.WriteByte(value, valueType)
		sb.WriteByte(value >> 8, STANDARD)
		sb.WriteByte(value >> 16, STANDARD)
		sb.WriteByte(value >> 24, STANDARD)
	}
}

func (sb *StreamBuffer) WriteLong(value int64, valueType ValueType, order ByteOrder) {
	switch order {
	case BIG:
		sb.WriteByte(int(value >> 56), STANDARD)
		sb.WriteByte(int(value >> 48), STANDARD)
		sb.WriteByte(int(value >> 40), STANDARD)
		sb.WriteByte(int(value >> 32), STANDARD)
		sb.WriteByte(int(value >> 24), STANDARD)
		sb.WriteByte(int(value >> 16), STANDARD)
		sb.WriteByte(int(value >> 8), STANDARD)
		sb.WriteByte(int(value), STANDARD)
	case LITTLE:
		sb.WriteByte(int(value), STANDARD)
		sb.WriteByte(int(value >> 8), STANDARD)
		sb.WriteByte(int(value >> 16), STANDARD)
		sb.WriteByte(int(value >> 24), STANDARD)
		sb.WriteByte(int(value >> 32), STANDARD)
		sb.WriteByte(int(value >> 40), STANDARD)
		sb.WriteByte(int(value >> 48), STANDARD)
		sb.WriteByte(int(value >> 56), STANDARD)
	}
}

func (sb *StreamBuffer) WriteString(str string) {
	for _, c := range []byte(str) {
		sb.WriteByte(int(c), STANDARD)
	}
	sb.WriteByte(10, STANDARD)
}

func (sb *StreamBuffer) SetAccessType(accessType AccessType) {
	sb.accessType = accessType
	switch accessType {
	case BIT_ACCESS:
		sb.bitPosition = sb.Buffer.Position * 8
	case BYTE_ACCESS:
		sb.Buffer.Position = (sb.bitPosition + 7) / 8
	}
}
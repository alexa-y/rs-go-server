package io

import (
	"errors"
	"rs-go-server/crypto"
	"strings"
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
	input bool
}

func NewOutBuffer(size int) *StreamBuffer {
	buffer := NewByteBuffer(size)
	sb := &StreamBuffer{size: size, Buffer: buffer}
	return sb
}

func NewInBuffer(buf *ByteBuffer) *StreamBuffer {
	sb := &StreamBuffer{Buffer: buf, input: true}
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
	for _, b := range buf.Buf {
		sb.WriteByte(int(b), STANDARD)
	}
}

func (sb *StreamBuffer) WriteBytesReverse(buf *ByteBuffer) {
	for i := buf.Cap() - 1; i >= 0; i-- {
		sb.WriteByte(int(buf.Buf[i]), STANDARD)
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

func (sb *StreamBuffer) ReadByte(valueType ValueType) byte {
	val, _ := sb.Buffer.Read()
	switch valueType {
	case A:
		val = val - 128
	case C:
		val = -val
	case S:
		val = 128 - val
	}
	return val
}

func (sb *StreamBuffer) ReadShort(valueType ValueType, order ByteOrder) uint16 {
	var val uint16
	switch order {
	case BIG:
		val |= uint16(sb.ReadByte(STANDARD)) << 8
		val |= uint16(sb.ReadByte(valueType))
	case LITTLE:
		val |= uint16(sb.ReadByte(valueType))
		val |= uint16(sb.ReadByte(STANDARD)) << 8
	}
	return val
}

func (sb *StreamBuffer) ReadInt(valueType ValueType, order ByteOrder) uint32 {
	var val uint32
	switch order {
	case BIG:
		val |= uint32(sb.ReadByte(STANDARD)) << 24
		val |= uint32(sb.ReadByte(STANDARD)) << 16
		val |= uint32(sb.ReadByte(STANDARD)) << 8
		val |= uint32(sb.ReadByte(valueType))
	case MIDDLE:
		val |= uint32(sb.ReadByte(STANDARD)) << 8
		val |= uint32(sb.ReadByte(valueType))
		val |= uint32(sb.ReadByte(STANDARD)) << 24
		val |= uint32(sb.ReadByte(STANDARD)) << 16
	case INVERSE_MIDDLE:
		val |= uint32(sb.ReadByte(STANDARD)) << 16
		val |= uint32(sb.ReadByte(STANDARD)) << 24
		val |= uint32(sb.ReadByte(valueType))
		val |= uint32(sb.ReadByte(STANDARD)) << 8
	case LITTLE:
		val |= uint32(sb.ReadByte(valueType))
		val |= uint32(sb.ReadByte(STANDARD)) << 8
		val |= uint32(sb.ReadByte(STANDARD)) << 16
		val |= uint32(sb.ReadByte(STANDARD)) << 24
	}
	return val
}

func (sb *StreamBuffer) ReadLong(valueType ValueType, order ByteOrder) uint64 {
	var val uint64
	switch order {
	case BIG:
		val |= uint64(sb.ReadByte(STANDARD)) << 56
		val |= uint64(sb.ReadByte(STANDARD)) << 48
		val |= uint64(sb.ReadByte(STANDARD)) << 40
		val |= uint64(sb.ReadByte(STANDARD)) << 32
		val |= uint64(sb.ReadByte(STANDARD)) << 24
		val |= uint64(sb.ReadByte(STANDARD)) << 16
		val |= uint64(sb.ReadByte(STANDARD)) << 8
		val |= uint64(sb.ReadByte(valueType))
	case LITTLE:
		val |= uint64(sb.ReadByte(valueType))
		val |= uint64(sb.ReadByte(STANDARD)) << 8
		val |= uint64(sb.ReadByte(STANDARD)) << 16
		val |= uint64(sb.ReadByte(STANDARD)) << 24
		val |= uint64(sb.ReadByte(STANDARD)) << 32
		val |= uint64(sb.ReadByte(STANDARD)) << 40
		val |= uint64(sb.ReadByte(STANDARD)) << 48
		val |= uint64(sb.ReadByte(STANDARD)) << 56
	}
	return val
}

func (sb *StreamBuffer) ReadString() string {
	builder := strings.Builder{}
	for {
		tmp := sb.ReadByte(STANDARD)
		if tmp == 10 {
			break
		} else {
			builder.WriteByte(tmp)
		}
	}
	return builder.String()
}

func (sb *StreamBuffer) ReadBytes(amount int, valueType ValueType) []byte {
	data := make([]byte, amount)
	for i := 0; i < amount; i++ {
		data[i] = sb.ReadByte(valueType)
	}
	return data
}

func (sb *StreamBuffer) ReadBytesReverse(amount int, valueType ValueType) []byte {
	data := make([]byte, amount)
	dataPosition := 0
	for i := sb.Buffer.Position + amount - 1; i >= sb.Buffer.Position; i-- {
		val, _ := sb.Buffer.Get(i)
		switch valueType {
		case A:
			val -= 128
		case C:
			val = -val
		case S:
			val = 128 - val
		}
		data[dataPosition] = val
		dataPosition++
	}
	return data
}
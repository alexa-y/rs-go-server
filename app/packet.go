package app

import "rs-go-server/io"

type Packet struct {
	ID byte
	Length byte
	Data *io.ByteBuffer
}
package app

import (
	"fmt"
	"rs-go-server/io"
)

func HandleButtonPacket(p *Player, packet *Packet) {
	buf := io.NewInBuffer(packet.Data)

	button := buf.ReadShort(io.STANDARD, io.BIG)
	fmt.Printf("Button: %d\n", button)
	switch button {
	case 9154:
		p.SendLogout()
	}
}
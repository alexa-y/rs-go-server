package app

import (
	"fmt"
	"rs-go-server/io"
)

func HandleButtonPacket(p *Player, packet *Packet) {
	buf := io.NewInBuffer(packet.Data)

	buttonBytes := buf.ReadBytes(2, io.STANDARD)
	button := HexToInt(buttonBytes)
	fmt.Printf("Button: %v\n", button)
	switch button {
	case 9154:
		p.SendLogout()
	}
}

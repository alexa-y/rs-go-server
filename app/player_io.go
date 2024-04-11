package app

import (
	"crypto/rand"
	"fmt"
	"rs-go-server/crypto"
	"rs-go-server/io"
	"strings"
)

// responsible for I/O for player

type UnexpectedPacketSizeError struct{ Received, Expected int }

func (e UnexpectedPacketSizeError) Error() string {
	return fmt.Sprintf("client: Unexpected Packet Size Error.  Received: %d, Expected: %d", e.Received, e.Expected)
}

type InvalidLoginRequestError struct{ Request byte }

func (e InvalidLoginRequestError) Error() string {
	return fmt.Sprintf("client: Invalid login request.  Request: %d", e.Request)
}

type InvalidClientVersionError struct{ Version uint16 }

func (e InvalidClientVersionError) Error() string {
	return fmt.Sprintf("client: Invalid client version.  Version: %d", e.Version)
}

func (p *Player) HandleIncomingData() error {
	incomingData := make([]byte, 8192)
	size, err := p.Socket.Read(incomingData)
	p.inBuffer.Compact()
	p.inBuffer.Append(incomingData[:size])
	p.inBuffer.Flip()

	if err != nil {
		fmt.Printf("Player incoming data error: %v", err)
		return err
	}
	buffer := io.NewInBuffer(p.inBuffer)

	if p.LoginStage != LOGGED_IN {
		return p.handleLogin(buffer)
	}

	for p.inBuffer.Remaining() > 0 {
		if p.PacketID == 0xFF {
			packetId, _ := p.inBuffer.Read()
			p.PacketID = packetId
			p.PacketID -= byte(p.Decryptor.Next())
			fmt.Printf("Packet ID: %d\n", p.PacketID)
		}

		if p.PacketLength == 0xFF {
			p.PacketLength = PACKET_SIZES[p.PacketID]
			if p.PacketLength == 0xFF {
				if p.inBuffer.Remaining() > 0 {
					packetLength, _ := p.inBuffer.Read()
					p.PacketLength = packetLength
				} else {
					return nil
				}
			}
		}

		if p.inBuffer.Remaining() >= int(p.PacketLength) {
			data := make([]byte, p.PacketLength)
			for i := range data {
				data[i], _ = p.inBuffer.Read()
			}
			packet := &Packet{p.PacketID, p.PacketLength, io.NewByteBufferWithBytes(data)}
			packet.Data.Flip()
			p.handlePacket(packet)

			p.PacketID = 0xFF
			p.PacketLength = 0xFF
		}
	}
	return nil
}

func (p *Player) handleLogin(buffer *io.StreamBuffer) error {
	switch p.LoginStage {
	case CONNECTED:
		if l := buffer.Remaining(); l < 2 {
			return UnexpectedPacketSizeError{Expected: 2, Received: l}
		}

		request := buffer.Read()
		buffer.Read() // name hash
		if request != 14 {
			return InvalidLoginRequestError{Request: request}
		}

		out := io.NewOutBuffer(17)
		out.WriteLong(0, io.STANDARD, io.BIG) // ignored by client
		out.WriteByte(0, io.STANDARD)         // response opcode, 0 for logging in
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		out.WriteBytes(io.NewByteBufferWithBytes(randBytes))
		err := p.Send(out)

		p.LoginStage = LOGGING_IN
		return err
	case LOGGING_IN:
		if l := buffer.Buffer.Remaining(); l < 2 {
			return UnexpectedPacketSizeError{Expected: 2, Received: l}
		}

		request, _ := buffer.Buffer.Read()
		if request != 16 && request != 18 {
			return InvalidLoginRequestError{Request: request}
		}

		blockLength, _ := buffer.Buffer.Read()
		if buffer.Buffer.Remaining() < int(blockLength) {
			buffer.Buffer.Flip()
			return nil
		}

		buffer.ReadByte(io.STANDARD) // magic ID

		clientVersion := buffer.ReadShort(io.STANDARD, io.BIG)
		if clientVersion != 317 {
			return InvalidClientVersionError{Version: clientVersion}
		}

		buffer.ReadByte(io.STANDARD) // high/low memory
		for i := 0; i < 9; i++ {     // CRC Keys
			buffer.ReadInt(io.STANDARD, io.BIG)
		}
		buffer.ReadByte(io.STANDARD) // RSA block length
		buffer.ReadByte(io.STANDARD) // RSA opcode
		buffer.ReadString()          // codebase

		clientHalf := buffer.ReadLong(io.STANDARD, io.BIG)
		serverHalf := buffer.ReadLong(io.STANDARD, io.BIG)
		isaacSeed := [...]uint32{uint32(clientHalf >> 32), uint32(clientHalf), uint32(serverHalf >> 32), uint32(serverHalf)}
		//p.Decryptor = crypto.NewISAACCipher(isaacSeed[:])
		p.Decryptor = crypto.NewMockISAACCipher(isaacSeed[:])
		for i := range isaacSeed {
			isaacSeed[i] += 50
		}
		//p.Encryptor = crypto.NewISAACCipher(isaacSeed[:])
		p.Encryptor = crypto.NewMockISAACCipher(isaacSeed[:])

		buffer.ReadInt(io.STANDARD, io.BIG) // user ID
		p.Username = strings.TrimSpace(buffer.ReadString())
		p.Password = []byte(buffer.ReadString())

		err := p.Login()
		if err != nil {
			return err
		}
		p.LoginStage = LOGGED_IN
	}
	return nil
}

func (p *Player) handlePacket(packet *Packet) {
	switch packet.ID {
	case 185: //button clicking
		HandleButtonPacket(p, packet)
	}
}

func (p *Player) SendLoginFrame() error {
	buffer := io.NewOutBuffer(3)
	buffer.WriteByte(2, io.STANDARD)
	buffer.WriteByte(0, io.STANDARD)
	buffer.WriteByte(0, io.STANDARD)
	return p.Send(buffer)
}

func (p *Player) SendMapRegion() error {
	buffer := io.NewOutBuffer(5)
	buffer.WriteHeader(p.Encryptor, 69)
	buffer.WriteShort(p.Position.RegionX()+6, io.A, io.BIG)
	buffer.WriteShort(p.Position.RegionY()+6, io.STANDARD, io.BIG)
	return p.Send(buffer)
}

func (p *Player) sendUpdate() error {
	out := io.NewOutBuffer(4096)
	block := io.NewOutBuffer(3072)

	out.WriteVariableShortPacketHeader(p.Encryptor, 81)
	out.SetAccessType(io.BIT_ACCESS)

	p.updateLocalPlayerMovement(out)
	if p.UpdateRequired {
		p.updateState(block)
	}

	out.WriteBits(8, 0)

	if block.Buffer.Position > 0 {
		out.WriteBits(11, 2047)
		out.SetAccessType(io.BYTE_ACCESS)
		out.WriteBytes(block.Buffer)
	} else {
		out.SetAccessType(io.BYTE_ACCESS)
	}

	out.FinishVariableShortPacketHeader()
	return p.Send(out)
}

func (p *Player) updateLocalPlayerMovement(buf *io.StreamBuffer) {
	if p.UpdateRequired {
		buf.WriteBit(true)
		buf.WriteBits(2, 3)
		buf.WriteBits(2, p.Position.Z)
		buf.WriteBit(true)
		buf.WriteBit(true)
		buf.WriteBits(7, p.Position.LocalY())
		buf.WriteBits(7, p.Position.LocalX())
	} else {
		buf.WriteBit(false)
	}
}

func (p *Player) updateState(buf *io.StreamBuffer) {
	const mask = 0x10
	buf.WriteByte(mask, io.STANDARD)
	p.appendAppearance(buf)
}

func (p *Player) appendAppearance(buf *io.StreamBuffer) {
	block := io.NewOutBuffer(128)
	block.WriteByte(0, io.STANDARD)
	block.WriteByte(0, io.STANDARD)

	// equipment
	block.WriteByte(0, io.STANDARD)
	block.WriteByte(0, io.STANDARD)
	block.WriteByte(0, io.STANDARD)
	block.WriteByte(0, io.STANDARD)
	block.WriteShort(0x100+18, io.STANDARD, io.BIG)
	block.WriteByte(0, io.STANDARD)
	block.WriteShort(0x100+26, io.STANDARD, io.BIG)
	block.WriteShort(0x100+36, io.STANDARD, io.BIG)
	block.WriteShort(0x100, io.STANDARD, io.BIG)
	block.WriteShort(0x100+33, io.STANDARD, io.BIG)
	block.WriteShort(0x100+42, io.STANDARD, io.BIG)
	block.WriteShort(0x100+10, io.STANDARD, io.BIG)

	// colors
	block.WriteByte(7, io.STANDARD)
	block.WriteByte(8, io.STANDARD)
	block.WriteByte(9, io.STANDARD)
	block.WriteByte(5, io.STANDARD)
	block.WriteByte(0, io.STANDARD)

	// animations
	block.WriteShort(808, io.STANDARD, io.BIG)
	block.WriteShort(0x337, io.STANDARD, io.BIG)
	block.WriteShort(819, io.STANDARD, io.BIG)
	block.WriteShort(0x334, io.STANDARD, io.BIG)
	block.WriteShort(0x335, io.STANDARD, io.BIG)
	block.WriteShort(0x336, io.STANDARD, io.BIG)
	block.WriteShort(824, io.STANDARD, io.BIG)

	block.WriteString(p.Username)
	block.WriteByte(3, io.STANDARD)
	block.WriteShort(0, io.STANDARD, io.BIG)

	buf.WriteByte(block.Buffer.Position, io.C)
	buf.WriteBytes(block.Buffer)
}

func (p *Player) SendSidebarInterface(idx, val int) {
	buf := io.NewOutBuffer(4)
	buf.WriteHeader(p.Encryptor, 71)
	buf.WriteShort(val, io.STANDARD, io.BIG)
	buf.WriteByte(idx, io.A)
	p.Send(buf)
}

func (p *Player) SendInventory() {
	buf := io.NewOutBuffer(256)
	buf.WriteVariableShortPacketHeader(p.Encryptor, 53)
	buf.WriteShort(3214, io.STANDARD, io.BIG)
	buf.WriteShort(len(p.Inventory), io.STANDARD, io.BIG)
	for _, item := range p.Inventory {
		if item.Amount > 254 {
			buf.WriteByte(255, io.STANDARD)
			buf.WriteInt(item.Amount, io.STANDARD, io.INVERSE_MIDDLE)
		} else {
			buf.WriteByte(item.Amount, io.STANDARD)
		}
		buf.WriteShort(item.ID+1, io.A, io.LITTLE)
	}
	buf.FinishVariableShortPacketHeader()
	p.Send(buf)
}

func (p *Player) SendLogout() {
	buf := io.NewOutBuffer(1)
	buf.WriteHeader(p.Encryptor, 109)
	p.Send(buf)
}

// func (p *Player) Send(buffer *io.ByteBuffer) error {
// 	_, err := p.Socket.Write(buffer.Buffer())
// 	return err
// }

func (p *Player) Send(buffer *io.StreamBuffer) error {
	_, err := buffer.WriteTo(p.Socket)
	return err
}

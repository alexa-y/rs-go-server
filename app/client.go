package app

import (
	"crypto/rand"
	"fmt"
	"rs-go-server/crypto"
	"rs-go-server/io"
	"strings"
)

// responsible for I/O for player

type UnexpectedPacketSizeError struct { Received, Expected int }
func (e UnexpectedPacketSizeError) Error() string {
	return fmt.Sprintf("client: Unexpected Packet Size Error.  Received: %d, Expected: %d", e.Received, e.Expected)
}

type InvalidLoginRequestError struct { Request byte }
func (e InvalidLoginRequestError) Error() string {
	return fmt.Sprintf("client: Invalid login request.  Request: %d", e.Request)
}

type InvalidClientVersionError struct { Version uint16 }
func (e InvalidClientVersionError) Error() string {
	return fmt.Sprintf("client: Invalid client version.  Version: %d", e.Version)
}

func (p *Player) HandleIncomingData() error {
	incomingData := make([]byte, 2048)
	size, err := p.Socket.Read(incomingData)
	p.inBuffer.Compact()
	p.inBuffer.Append(incomingData[:size])
	p.inBuffer.Flip()

	if err != nil {
		fmt.Println("Player incoming data error")
		return err
	}
	buffer := io.NewInBuffer(p.inBuffer)

	if p.LoginStage != LOGGED_IN {
		return p.handleLogin(buffer)
	}
	return nil
}

func (p *Player) handleLogin(buffer *io.StreamBuffer) error {
	switch p.LoginStage {
	case CONNECTED:
		if l := buffer.Buffer.Remaining(); l < 2 {
			return UnexpectedPacketSizeError{ Expected: 2, Received: l }
		}

		request, _ := buffer.Buffer.Read()
		buffer.Buffer.Read() // name hash
		if request != 14 {
			return InvalidLoginRequestError{Request: request}
		}

		out := io.NewOutBuffer(17)
		out.WriteLong(0, io.STANDARD, io.BIG) // ignored by client
		out.WriteByte(0, io.STANDARD) // response opcode, 0 for logging in
		randBytes := make([]byte, 8)
		rand.Read(randBytes)
		out.WriteBytes(&io.ByteBuffer{Buf: randBytes})
		err := p.Send(out.Buffer)

		p.LoginStage = LOGGING_IN
		return err
	case LOGGING_IN:
		if l := buffer.Buffer.Remaining(); l < 2 {
			return UnexpectedPacketSizeError{ Expected: 2, Received: l }
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
		for i := 0; i < 9; i++ { // CRC Keys
			buffer.ReadInt(io.STANDARD, io.BIG)
		}
		buffer.ReadByte(io.STANDARD) // RSA block length
		buffer.ReadByte(io.STANDARD) // RSA opcode
		buffer.ReadString() // codebase

		clientHalf := buffer.ReadLong(io.STANDARD, io.BIG)
		serverHalf := buffer.ReadLong(io.STANDARD, io.BIG)
		isaacSeed := [...]uint32{ uint32(clientHalf >> 32), uint32(clientHalf), uint32(serverHalf >> 32), uint32(serverHalf) }
		p.Decryptor = crypto.NewISAACCipher(isaacSeed[:])
		for i, _ := range isaacSeed {
			isaacSeed[i] += 50
		}
		p.Encryptor = crypto.NewISAACCipher(isaacSeed[:])

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

func (p *Player) SendLoginFrame() error {
	buffer := io.NewOutBuffer(3)
	buffer.WriteByte(2, io.STANDARD)
	buffer.WriteByte(0, io.STANDARD)
	buffer.WriteByte(0, io.STANDARD)
	return p.Send(buffer.Buffer)
}

func (p *Player) SendMapRegion() error {
	buffer := io.NewOutBuffer(5)
	buffer.WriteHeader(p.Encryptor, 69)
	buffer.WriteShort(p.Position.RegionX() + 6, io.A, io.BIG)
	buffer.WriteShort(p.Position.RegionY() + 6, io.STANDARD, io.BIG)
	return p.Send(buffer.Buffer)
}

func (p *Player) Send(buffer *io.ByteBuffer) error {
	_, err := p.Socket.Write(buffer.Buf)
	return err
}
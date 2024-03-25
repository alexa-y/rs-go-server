package app

import (
	"fmt"
	"net"
	"rs-go-server/crypto"
	"rs-go-server/io"
)

const (
	CONNECTED = 0
	LOGGING_IN = 1
	LOGGED_IN = 2
)

type Player struct {
	Socket *net.TCPConn
	LoginStage int
	UpdateRequired bool
	Connected bool
	Username string
	Password []byte
	inBuffer *io.ByteBuffer
	Encryptor *crypto.ISAACCipher
	Decryptor *crypto.ISAACCipher
	Position *Position
}

func NewPlayer(socket *net.TCPConn) *Player {
	player := &Player{ Socket: socket, Connected: true, inBuffer: io.NewByteBuffer(512), UpdateRequired: true }
	player.Position = &Position{X: 3222, Y: 3218}
	return player
}

func (p *Player) Process() {
	err := p.HandleIncomingData()
	if err != nil {
		fmt.Println(err)
		p.Connected = false
		p.Socket.Close()
		return
	}
}

func (p *Player) Update() {
	p.sendUpdate()
	p.UpdateRequired = false
}

func (p *Player) Login() error {
	p.SendLoginFrame()
	p.SendMapRegion()
	return nil
}
package app

import (
	"fmt"
	"net"
	"rs-go-server/crypto"
	"rs-go-server/io"
)

const (
	CONNECTED  = 0
	LOGGING_IN = 1
	LOGGED_IN  = 2
)

type Player struct {
	Socket         *net.TCPConn
	LoginStage     int
	UpdateRequired bool
	Connected      bool
	Username       string
	Password       []byte
	inBuffer       *io.ByteBuffer
	Encryptor      crypto.Cipher
	Decryptor      crypto.Cipher
	Position       *Position
	Inventory      ItemContainer
	PacketID       byte
	PacketLength   byte
}

func NewPlayer(socket *net.TCPConn) *Player {
	player := &Player{
		Socket:         socket,
		Connected:      true,
		inBuffer:       io.NewByteBuffer(512),
		UpdateRequired: true,
		PacketID:       0xFF,
		PacketLength:   0xFF,
	}
	player.Position = &Position{X: 3222, Y: 3218}
	player.Inventory = NewItemContainer(28)
	for _, i := range [...]int{1038, 1040, 1042, 1044, 1046, 1048} {
		player.Inventory.Add(&Item{i, 1})
	}
	return player
}

func (p *Player) Process() error {
	err := p.HandleIncomingData()
	if err != nil {
		fmt.Println(err)
		p.Connected = false
		p.Socket.Close()
	}
	return err
}

func (p *Player) Update() {
	p.sendUpdate()
	p.UpdateRequired = false
}

func (p *Player) Login() error {
	p.SendLoginFrame()
	p.SendMapRegion()
	p.SendInventory()
	p.SendSidebarInterface(0, 5855)
	p.SendSidebarInterface(1, 3917)
	p.SendSidebarInterface(2, 638)
	p.SendSidebarInterface(3, 3213)
	p.SendSidebarInterface(4, 1644)
	p.SendSidebarInterface(5, 5608)
	p.SendSidebarInterface(6, 1151)
	p.SendSidebarInterface(8, 5065)
	p.SendSidebarInterface(9, 5715)
	p.SendSidebarInterface(10, 2449)
	p.SendSidebarInterface(11, 904)
	p.SendSidebarInterface(12, 147)
	p.SendSidebarInterface(13, 962)
	return nil
}

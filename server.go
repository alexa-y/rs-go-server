package main

import (
	"fmt"
	"net"
	"rs-go-server/app"
	"time"
)

const (
	Port        int = 43594
	MaxPlayers      = 2000
	CycleMillis     = 600
)

var (
	Players = make([]*app.Player, MaxPlayers)
)

func main() {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: Port})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening on %v\n", listener.Addr())
	go UpdatePlayers()
	for {
		connection, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(connection.RemoteAddr())
		if slot := NextPlayerSlot(); slot >= 0 {
			Players[NextPlayerSlot()] = app.NewPlayer(connection)
		}
	}
}

func UpdatePlayers() {
	for {
		cycleStart := time.Now()
		for i, p := range Players {
			if p == nil {
				continue
			}
			if p.Connected {
				err := p.Process()
				if err != nil {
					Players[i] = nil
					continue
				}
			}
		}
		for _, p := range Players {
			if p == nil {
				continue
			}
			if p.Connected && p.LoginStage == app.LOGGED_IN {
				p.Update()
			}
		}
		time.Sleep(time.Now().Sub(cycleStart) + CycleMillis*time.Millisecond)
	}
}

func NextPlayerSlot() int {
	for i, p := range Players {
		if p == nil || !p.Connected {
			return i
		}
	}
	return -1
}

package main

import (
	"fmt"
	"rs-go-server/io"
)

var (
	Port int
)

func main() {
	//cipher := crypto.NewISAACCipher([]uint32{1, 2, 3 })

	bb := io.NewByteBuffer(10)
	for i := 0; i < 15; i++ {
		err := bb.Write(byte(i))
		if err != nil {
			fmt.Println(err.Error())
			bb.Resize(bb.Cap() + 1)
			bb.Write(byte(i))
		}
	}
	fmt.Println(bb.Remaining())
	fmt.Println(bb)
}

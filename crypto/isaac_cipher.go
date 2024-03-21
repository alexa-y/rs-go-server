package crypto

const Size uint32 = 256

type ISAACCipher struct {
	a, b, c uint32
	memory  [256]uint32
	results [256]uint32
	count   uint32
}

func NewISAACCipher(seed []uint32) (z *ISAACCipher) {
	z = &ISAACCipher{}
	for i, r := range seed {
		if i == 256 {
			break
		}
		z.results[i] = r
	}
	z.randInit()
	return
}

func (z *ISAACCipher) isaac() {
	z.c++
	z.b += z.c
	for i, x := range z.memory {
		switch i % 4 {
		case 0:
			z.a = z.a ^ z.a<<13
		case 1:
			z.a = z.a ^ z.a>>6
		case 2:
			z.a = z.a ^ z.a<<2
		case 3:
			z.a = z.a ^ z.a>>16
		}
		z.a += z.memory[(i+128)%256]
		y := z.memory[x>>2%256] + z.a + z.b
		z.memory[i] = y
		z.b = z.memory[y>>10%256] + x
		z.results[i] = z.b
	}
}

func (z *ISAACCipher) randInit() {
	const gold = uint32(0x9e3779b9)
	a := [8]uint32{gold, gold, gold, gold, gold, gold, gold, gold}
	mix1 := func(i int, v uint32) {
		a[i] ^= v
		a[(i+3)%8] += a[i]
		a[(i+1)%8] += a[(i+2)%8]
	}
	mix := func() {
		mix1(0, a[1]<<11)
		mix1(1, a[2]>>2)
		mix1(2, a[3]<<8)
		mix1(3, a[4]>>16)
		mix1(4, a[5]<<10)
		mix1(5, a[6]>>4)
		mix1(6, a[7]<<8)
		mix1(7, a[0]>>9)
	}
	for i := 0; i < 4; i++ {
		mix()
	}
	for i := 0; i < 256; i += 8 {
		for j, rj := range z.results[i : i+8] {
			a[j] += rj
		}
		mix()
		for j, aj := range a {
			z.memory[i+j] = aj
		}
	}
	for i := 0; i < 256; i += 8 {
		for j, mj := range z.memory[i : i+8] {
			a[j] += mj
		}
		mix()
		for j, aj := range a {
			z.memory[i+j] = aj
		}
	}
	z.isaac()
	z.count = Size
}

func (z *ISAACCipher) Next() uint32 {
	if z.count == 0 {
		z.isaac()
		z.count = Size - 1
	} else {
		z.count--
	}
	return z.results[z.count]
}
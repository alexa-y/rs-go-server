package app

type Position struct {
	X, Y, Z int
}

func (p *Position) RegionX() int {
	return (p.X >> 3) - 6
}

func (p *Position) RegionY() int {
	return (p.Y >> 3) - 6
}

func (p *Position) LocalX() int {
	return p.X - 8 * p.RegionX()
}

func (p *Position) LocalY() int {
	return p.Y - 8 * p.RegionY()
}
package svg

type Point struct {
	X, Y float64
}

func (p Point) Add(other Point) Point {
	p.X += other.X
	p.Y += other.Y

	return p
}

func (p *Point) Reset() {
	p.X = 0
	p.Y = 0
}

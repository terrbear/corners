package rpc

type Point struct {
	X, Y int
}

func (p *Point) Values() (int, int) {
	return p.X, p.Y
}

func (p *Point) Offset(x, y int) Point {
	return Point{p.X + x, p.Y + y}
}

func (p *Point) Distance(to *Point) float64 {
	absX := to.X - p.X
	if absX < 0 {
		absX = -absX
	}

	absY := to.Y - p.Y
	if absY < 0 {
		absY = -absY
	}

	return float64(absX + absY)
}

package corners

import (
	"math/rand"
	"time"
)

type Point struct {
	x, y int
}

func (p *Point) Values() (int, int) {
	return p.x, p.y
}

func (p *Point) Offset(x, y int) Point {
	return Point{p.x + x, p.y + y}
}

type tileTracker struct {
	size int
	grid [][]string
}

func newTileTracker(size int) tileTracker {
	tt := tileTracker{
		size: size,
	}

	tt.grid = make([][]string, tt.size)
	for x := range tt.grid {
		tt.grid[x] = make([]string, tt.size)
		for y := range tt.grid[x] {
			tt.grid[x][y] = "free"
		}
	}

	return tt
}

func (tt *tileTracker) findAllFreePoints() []Point {
	var points []Point

	for x := range tt.grid {
		for y := range tt.grid[x] {
			if tt.grid[x][y] == "free" {
				points = append(points, Point{x, y})
			}
		}
	}

	return points
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (tt *tileTracker) pickRandomFreePoint() Point {
	freePoints := tt.findAllFreePoints()

	// todo: decide how to handle this gracefully
	if len(freePoints) == 0 {
		panic("there are no more free points!")
	}

	return freePoints[rand.Intn(len(freePoints))]
}

func (tt *tileTracker) getAt(point Point) string {
	return tt.grid[point.x][point.y]
}

func (tt *tileTracker) setAt(point Point, value string) {
	if point.x < 0 || point.y < 0 {
		return
	}
	if point.x >= tt.size || point.y >= tt.size {
		return
	}

	tt.grid[point.x][point.y] = value
}

func (tt *tileTracker) occupyRandomPoint(object string) Point {
	point := tt.pickRandomFreePoint()

	tt.occupyPoint(point, object)

	return point
}

func (tt *tileTracker) occupyPoint(point Point, object string) {
	// todo: decide how to handle this gracefully
	if tt.getAt(point) != "free" {
		panic("space is not free!")
	}

	tt.setAt(point, "occupied")

	// generators cannot be right next to each other
	if object == "generator" {
		tt.setAt(point.Offset(-1, -1), "claimed")
		tt.setAt(point.Offset(+0, -1), "claimed")
		tt.setAt(point.Offset(+1, -1), "claimed")
		tt.setAt(point.Offset(-1, +0), "claimed")

		tt.setAt(point.Offset(+1, +0), "claimed")
		tt.setAt(point.Offset(-1, +1), "claimed")
		tt.setAt(point.Offset(+0, +1), "claimed")
		tt.setAt(point.Offset(+1, +1), "claimed")
	}
}

type Options struct {
	startingPoints     [][2]int
	numberOfGenerators int
	numberOfWalls      int
}

func GenerateRandomMap(options Options) Map {
	tt := newTileTracker(16)
	ma := Map{
		Name:           "random",
		Size:           16,
		StartingPoints: options.startingPoints,
		Overrides:      []MapOverride{},
	}

	for _, point := range ma.StartingPoints {
		x, y := point[0], point[1]

		tt.occupyPoint(Point{x, y}, "generator")

		ma.Overrides = append(ma.Overrides, MapOverride{
			X:         x,
			Y:         y,
			Generator: true,
			Armies:    10,
		})
	}

	for i := 0; i < options.numberOfGenerators; i++ {
		point := tt.occupyRandomPoint("generator")
		x, y := point.Values()

		ma.Overrides = append(ma.Overrides, MapOverride{
			X:         x,
			Y:         y,
			Generator: true,
			Armies:    10,
		})
	}

	for i := 0; i < options.numberOfWalls; i++ {
		point := tt.occupyRandomPoint("wall")
		x, y := point.Values()

		ma.Overrides = append(ma.Overrides, MapOverride{
			X:         x,
			Y:         y,
			Generator: false,
			Armies:    100,
		})
	}

	return ma
}

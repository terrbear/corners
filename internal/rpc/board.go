package rpc

import (
	"fmt"
	"math"
	"time"
)

// Board represents the game board.
type Board struct {
	Tiles    [][]*Tile `json:"tiles"`
	size     int
	bluebase *Tile
	redbase  *Tile
}

// NewBoard generates a new Board with giving a size.
func NewBoard(size int) *Board {
	tiles := make([][]*Tile, size)
	for x := range tiles {
		tiles[x] = make([]*Tile, size)
		for y := range tiles[x] {
			tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1})
		}
	}

	bluebase := NewTile(&TileParams{generator: true, team: 1})
	redbase := NewTile(&TileParams{generator: true, team: 2, x: size - 1, y: size - 1})
	tiles[0][0] = bluebase
	tiles[size-1][size-1] = redbase

	return &Board{
		size:     size,
		Tiles:    tiles,
		redbase:  redbase,
		bluebase: bluebase,
	}
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	for col := range b.Tiles {
		for row := range b.Tiles[col] {
			f(col, row, b.Tiles[col][row])
		}
	}

	return nil
}

// TODO make this try to follow a diagonal
func movementVector(from, to *Tile) (int, int) {
	if from.X < to.X {
		return 1, 0
	} else if from.X > to.X {
		return -1, 0
	}

	if from.Y < to.Y {
		return 0, 1
	} else if from.Y > to.Y {
		return 0, -1
	}
	return 0, 0
}

type Transfer struct {
	armies int
	from   *Tile
	to     *Tile
}

var transferRate = 500 * time.Millisecond

// var transferRate = time.Millisecond

func (b *Board) runTransfer(t *Transfer) {
	ticker := time.NewTicker(transferRate)

	for {
		if t.armies <= 1 {
			return
		}
		x, y := movementVector(t.from, t.to)
		if x == 0 && y == 0 {
			return
		}

		targetX, targetY := t.from.X+x, t.from.Y+y

		// fmt.Printf("moving %d armies from %d,%d to %d,%d\n", t.armies, t.from.x, t.from.y, targetX, targetY)

		dest := b.Tiles[targetX][targetY]

		if dest.Team != 0 && dest.Team != t.from.Team {
			fmt.Println("calling attack!")
			t.from.attack(dest)
		}

		if dest.Team == 0 || dest.Team == t.from.Team || dest.Armies == 0 {
			t.armies = t.from.take(t.armies)
			dest.add(t.from, t.armies)
			t.from = dest
		} else {
			fmt.Println("grabbing low bar from armies")
			t.armies = int(math.Min(float64(t.from.Armies), float64(t.armies)))
		}

		if t.from == t.to {
			return
		}

		fmt.Println("sleeping")
		<-ticker.C
	}
}

func (b *Board) Transfer(source, dest *Tile) {
	if source == dest {
		return
	}

	armies := source.Armies

	if armies == 0 {
		return
	}

	fmt.Printf("source offered %d armies\n", armies)

	t := Transfer{
		armies: armies,
		from:   source,
		to:     dest,
	}

	go b.runTransfer(&t)
}

// Update updates the board state.
func (b *Board) Tick() error {
	red, blue := 0, 0

	b.bluebase.resources = 0
	b.redbase.resources = 0

	b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.Team == 1 {
			blue += t.resources
		} else if t.Team == 2 {
			red += t.resources
		}
		return nil
	})

	// TODO lock this
	b.bluebase.resources = blue
	b.redbase.resources = red

	return nil
}

package rpc

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Board represents the game board.
type Board struct {
	Tiles [][]*Tile `json:"tiles"`
	size  int
	lock  sync.Mutex
}

type PlayerID string

const NeutralPlayer PlayerID = "neutral"

// NewBoard generates a new Board with giving a size.
func NewBoard(playerIDs []PlayerID, size int) *Board {
	tiles := make([][]*Tile, size)
	for x := range tiles {
		tiles[x] = make([]*Tile, size)
		for y := range tiles[x] {
			tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1, playerID: NeutralPlayer})
		}
	}

	positions := [][]int{{0, 0}, {size - 1, size - 1}, {0, size - 1}, {size - 1, 0}}

	for i := range positions {
		position := positions[i]
		playerID := NeutralPlayer
		if len(playerIDs) > i {
			playerID = playerIDs[i]
		}

		tiles[position[0]][position[1]] = NewTile(&TileParams{playerID: playerID, generator: true, x: position[0], y: position[1]})
	}

	b := &Board{
		size:  size,
		Tiles: tiles,
	}

	go func() {
		t := time.NewTicker(time.Second)
		for {
			b.Tick()
			<-t.C
		}
	}()

	return b
}

func (b *Board) Start() {
	b.forEach(0, 0, func(x, y int, tile *Tile) error {
		tile.Start()
		return nil
	})
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	for col := range b.Tiles {
		for row := range b.Tiles[col] {
			f(col, row, b.Tiles[col][row])
		}
	}

	return nil
}

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

		if dest.PlayerID != t.from.PlayerID && dest.Armies > 0 {
			fmt.Println("calling attack!")
			t.from.attack(dest)
		}

		if dest.PlayerID == t.from.PlayerID || dest.Armies == 0 {
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

func (b *Board) Transfer(playerID PlayerID, source, dest *Tile) {
	b.lock.Lock()
	defer b.lock.Unlock()

	if source == dest {
		return
	}

	if source.PlayerID != playerID {
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
	b.lock.Lock()
	defer b.lock.Unlock()

	resources := make(map[PlayerID]int)

	b.forEach(0, 0, func(x, y int, t *Tile) error {
		resources[t.PlayerID]++
		return nil
	})

	b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.generator && t.PlayerID != NeutralPlayer {
			t.resources = resources[t.PlayerID]
		}
		return nil
	})

	return nil
}

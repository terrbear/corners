package corners

import (
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"terrbear.io/corners/internal/rpc"
)

type Board struct {
	Tiles [][]*Tile
	lock  sync.Mutex
}

func NewBoard(playerIDs []rpc.PlayerID, size int) *Board {
	tiles := make([][]*Tile, size)
	for x := range tiles {
		tiles[x] = make([]*Tile, size)
		for y := range tiles[x] {
			tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1, playerID: rpc.NeutralPlayer})
		}
	}

	positions := [][]int{{0, 0}, {size - 1, size - 1}, {0, size - 1}, {size - 1, 0}}

	for i := range positions {
		position := positions[i]
		playerID := rpc.NeutralPlayer
		if len(playerIDs) > i {
			playerID = playerIDs[i]
		}

		tiles[position[0]][position[1]] = NewTile(&TileParams{playerID: playerID, generator: true, x: position[0], y: position[1]})
	}

	b := &Board{
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

func (b *Board) Start() error {
	return b.forEach(0, 0, func(x, y int, tile *Tile) error {
		tile.Start()
		return nil
	})
}

func (b *Board) ToRPCBoard() *rpc.Board {
	b.lock.Lock()
	defer b.lock.Unlock()

	board := &rpc.Board{
		Tiles: make([][]rpc.Tile, len(b.Tiles)),
	}

	for x := range b.Tiles {
		board.Tiles[x] = make([]rpc.Tile, len(b.Tiles[x]))
		for y := range b.Tiles[x] {
			board.Tiles[x][y] = b.Tiles[x][y].ToRPCTile()
		}
	}

	return board
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	for col := range b.Tiles {
		for row := range b.Tiles[col] {
			err := f(col, row, b.Tiles[col][row])
			if err != nil {
				return err
			}
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

		log.Tracef("moving %d armies from %d,%d to %d,%d\n", t.armies, t.from.X, t.from.Y, targetX, targetY)

		dest := b.Tiles[targetX][targetY]

		if dest.PlayerID != t.from.PlayerID && dest.Armies > 0 {
			log.Trace("attacking!")
			t.from.attack(dest)
		}

		if dest.PlayerID == t.from.PlayerID || dest.Armies == 0 {
			t.armies = t.from.moveTo(dest, t.armies)
			t.from = dest
		} else {
			log.Trace("grabbing low bar from armies")
			t.armies = int(math.Min(float64(t.from.Armies), float64(t.armies)))
		}

		if t.from == t.to {
			return
		}

		<-ticker.C
	}
}

func (b *Board) Transfer(playerID rpc.PlayerID, source, dest *Tile) {
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

	log.Tracef("source offered %d armies\n", armies)

	t := Transfer{
		armies: armies,
		from:   source,
		to:     dest,
	}

	go b.runTransfer(&t)
}

func (b *Board) Tick() {
	b.lock.Lock()
	defer b.lock.Unlock()

	resources := make(map[rpc.PlayerID]int)

	err := b.forEach(0, 0, func(x, y int, t *Tile) error {
		resources[t.PlayerID]++
		return nil
	})
	if err != nil {
		log.WithError(err).Error("error adding resources")
	}

	err = b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.generator && t.PlayerID != rpc.NeutralPlayer {
			t.resources = resources[t.PlayerID]
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Error("error adding resources to generators")
	}
}

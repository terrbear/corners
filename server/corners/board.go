package corners

import (
	"encoding/json"
	"math"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"terrbear.io/corners/internal/rpc"
)

type Board struct {
	Tiles  [][]*Tile
	Winner *rpc.PlayerID
	lock   sync.Mutex
	done   chan bool
}

type MapOverride struct {
	XR        []int
	YR        []int
	X         int
	Y         int
	Generator bool
	Armies    int
}

type Map struct {
	Name           string
	Size           int
	StartingPoints [][2]int
	Overrides      []MapOverride
}

func loadMap(name string) Map {
	if name == "random" {
		return GenerateRandomMap(Options{
			startingPoints: [][2]int{
				{02, 02},
				{13, 13},
				{13, 02},
				{02, 13},
			},
			numberOfGenerators: 0,
			numberOfWalls:      0,
		})
	}

	m, err := os.ReadFile(name + ".json")
	if err != nil {
		panic(err)
	}

	var setup Map
	err = json.Unmarshal(m, &setup)
	if err != nil {
		panic(err)
	}

	return setup
}

// NewBoard loads a board configuration from a given mapname and starts with the given playerIDs. Will
// start ticking immediately, but the tiles only start ticking once you call Start(). TODO is that necessary?
func NewBoard(mapName string, playerIDs []rpc.PlayerID) *Board {
	m := loadMap(mapName)

	tiles := make([][]*Tile, m.Size)
	for x := range tiles {
		tiles[x] = make([]*Tile, m.Size)
		for y := range tiles[x] {
			tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1, playerID: rpc.NeutralPlayer})
		}
	}

	for i, position := range m.StartingPoints {
		playerID := rpc.NeutralPlayer
		if len(playerIDs) > i {
			playerID = playerIDs[i]
		}

		tiles[position[0]][position[1]] = NewTile(&TileParams{playerID: playerID, resources: 1, x: position[0], y: position[1]})
	}

	for _, o := range m.Overrides {
		if len(o.XR) > 0 {
			for x := o.XR[0]; x <= o.XR[1]; x++ {
				for y := o.YR[0]; y <= o.YR[1]; y++ {
					tiles[x][y].generator = o.Generator
					tiles[x][y].Armies = o.Armies
				}
			}
		} else {
			tiles[o.X][o.Y].generator = o.Generator
			tiles[o.X][o.Y].Armies = o.Armies
		}
	}

	b := &Board{
		Tiles: tiles,
		done:  make(chan bool, 1),
	}

	go b.runTicker()

	return b
}

func (b *Board) Start() error {
	return b.forEach(0, 0, func(x, y int, tile *Tile) error {
		tile.Start()
		return nil
	})
}

func (b *Board) isVisible(x, y int, playerID rpc.PlayerID) bool {
	t := b.Tiles[x][y]
	if t.PlayerID == playerID {
		return true
	}
	leftX := int(math.Max(float64(x-1), 0))
	rightX := int(math.Min(float64(x+1), float64(len(b.Tiles)-1)))
	topY := int(math.Max(float64(y-1), 0))
	bottomY := int(math.Min(float64(y+1), float64(len(b.Tiles[0])-1)))

	// You can see the eight squares around you, so check each one
	return b.Tiles[leftX][y].PlayerID == playerID ||
		b.Tiles[leftX][topY].PlayerID == playerID ||
		b.Tiles[leftX][bottomY].PlayerID == playerID ||
		b.Tiles[x][topY].PlayerID == playerID ||
		b.Tiles[x][bottomY].PlayerID == playerID ||
		b.Tiles[rightX][y].PlayerID == playerID ||
		b.Tiles[rightX][topY].PlayerID == playerID ||
		b.Tiles[rightX][bottomY].PlayerID == playerID
}

func (b *Board) ToRPCBoard(player rpc.PlayerID) *rpc.Board {
	b.lock.Lock()
	defer b.lock.Unlock()

	board := &rpc.Board{
		Size:   len(b.Tiles),
		Tiles:  make([]rpc.Tile, len(b.Tiles)*len(b.Tiles[0])),
		Winner: b.Winner,
	}

	for x := range b.Tiles {
		for y := range b.Tiles[x] {
			board.Tiles = append(board.Tiles, b.Tiles[x][y].ToRPCTile(b.isVisible(x, y, player)))
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

func (b *Board) runTicker() {
	t := time.NewTicker(time.Second)
	for {
		b.tick()
		select {
		case <-t.C:
		case <-b.done:
			return
		}
	}
}

func (b *Board) tick() {
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

	generators := make(map[rpc.PlayerID]int)
	err = b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.generator {
			generators[t.PlayerID]++
			if t.PlayerID != rpc.NeutralPlayer {
				t.resources = resources[t.PlayerID]
			}
		}
		return nil
	})

	if len(generators) == 1 {
		// Only one player controls the generators; gg
		for w := range generators {
			b.Winner = &w
			b.done <- true
		}
	}

	if err != nil {
		log.WithError(err).Error("error adding resources to generators")
	}
}

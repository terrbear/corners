package corners

import (
	"image"
	"image/color"
	"sync"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	log "github.com/sirupsen/logrus"
	"terrbear.io/corners/internal/rpc"
)

// Board represents the game board.
type Board struct {
	selectedX int
	selectedY int
	targetX   int
	targetY   int

	size     int
	tiles    [][]*Tile
	offset   image.Point
	playerID rpc.PlayerID
	init     sync.Once
	lock     sync.Mutex

	client       *rpc.Client
	boardUpdates chan rpc.Board
	board        rpc.Board
}

func (b *Board) processBoardUpdates() {
	for board := range b.boardUpdates {
		b.lock.Lock()
		b.initBoard(board)
		b.board = board
		b.lock.Unlock()
	}
}

func (b *Board) initBoard(board rpc.Board) {
	b.init.Do(func() {
		log.Debug("initializing board with size: ", len(board.Tiles))
		tiles := make([][]*Tile, len(board.Tiles))
		for x := range tiles {
			tiles[x] = make([]*Tile, len(board.Tiles))
			for y := range tiles[x] {
				tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1})
			}
		}

		b.tiles = tiles
		b.size = len(board.Tiles)
	})
}

func NewBoard() *Board {
	b := &Board{
		playerID:     rpc.PlayerID(uuid.New().String()),
		boardUpdates: make(chan rpc.Board),
	}

	go b.processBoardUpdates()

	b.client = rpc.NewClient(b.playerID, b.boardUpdates)

	return b
}

func (b *Board) translate(mouse *image.Point) (int, int) {
	if mouse == nil {
		return -1, -1
	}
	p := mouse.Sub(b.offset)
	if p.X < 0 || p.Y < 0 {
		return -1, -1
	}
	x := p.X / (tileSize + tileMargin)
	y := p.Y / (tileSize + tileMargin)
	return x, y
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	if len(b.board.Tiles) == 0 {
		return nil
	}
	for col := range b.tiles {
		for row := range b.tiles[col] {
			b.tiles[row][col].tile = b.board.Tiles[row][col]
			err := f(col, row, b.tiles[col][row])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Board) transfer(sourceX, sourceY, targetX, targetY int) {
	b.client.SendCommand(rpc.Command{
		SelectedX: sourceX,
		SelectedY: sourceY,
		TargetX:   targetX,
		TargetY:   targetY,
	})
}

func (b *Board) Update(input *Input) error {
	clickedX, clickedY := b.translate(input.LeftMouse())
	targetX, targetY := b.translate(input.RightMouse())

	if clickedX >= 0 && clickedY >= 0 {
		b.selectedX = clickedX
		b.selectedY = clickedY
		log.Debugf("selected: %d, %d\n", clickedX, clickedY)
	}

	if b.selectedX >= 0 && b.selectedY >= 0 && targetX >= 0 && targetY >= 0 {
		log.Debugf("target: %d, %d\n", targetX, targetY)
		go func() {
			b.transfer(b.selectedX, b.selectedY, targetX, targetY)
		}()
	}

	return nil
}

// Size returns the board size.
func (b *Board) Size() (int, int) {
	x := b.size*tileSize + (b.size+1)*tileMargin
	y := x
	return x, y
}

// Draw draws the board to the given boardImage.
func (b *Board) Draw(boardImage *ebiten.Image) {
	boardImage.Fill(frameColor)

	for j := 0; j < b.size; j++ {
		for i := 0; i < b.size; i++ {
			op := &ebiten.DrawImageOptions{}
			x := i*tileSize + (i+1)*tileMargin
			y := j*tileSize + (j+1)*tileMargin
			cr, cg, cb, ca := colorToScale(color.NRGBA{0xee, 0xe4, 0xda, 0x59})
			op.GeoM.Translate(float64(x), float64(y))
			op.ColorM.Scale(cr, cg, cb, ca)
			boardImage.DrawImage(tileImage, op)
		}
	}
	if len(b.tiles) > 0 {
		err := b.forEach(0, 0, func(x, y int, t *Tile) error {
			t.Draw(x, y, boardImage, &TileDrawParams{
				boardPlayerID: b.playerID,
				targeted:      x == b.targetX && y == b.targetY,
				selected:      x == b.selectedX && y == b.selectedY,
			})
			return nil
		})
		if err != nil {
			log.WithError(err).Error("couldn't draw board")
		}
	}
}

package corners

import (
	"image"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	ScreenWidth  = 800
	ScreenHeight = 800
	boardSize    = 8
)

// Game represents a game state.
type Game struct {
	input      *Input
	board      *Board
	boardImage *ebiten.Image
}

// NewGame generates a new Game object.
func NewGame() (*Game, error) {
	g := &Game{
		input: NewInput(),
	}
	g.board = NewBoard(boardSize)
	return g, nil
}

// Layout implements ebiten.Game's Layout.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// Update updates the current game state.
func (g *Game) Update() error {
	g.input.Update()
	if err := g.board.Update(g.input); err != nil {
		return err
	}
	return nil
}

// Draw draws the current game to the given screen.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.boardImage == nil {
		w, h := g.board.Size()
		g.boardImage = ebiten.NewImage(w, h)
	}
	screen.Fill(backgroundColor)
	op := &ebiten.DrawImageOptions{}
	sw, sh := screen.Size()
	bw, bh := g.boardImage.Size()
	x := (sw - bw) / 2
	y := (sh - bh) / 2
	g.board.offset = image.Point{x, y}
	g.board.Draw(g.boardImage)
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(g.boardImage, op)
}

package corners

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// Board represents the game board.
type Board struct {
	size    int
	topLeft *Tile
}

// NewBoard generates a new Board with giving a size.
func NewBoard(size int) *Board {
	topLeft := NewTile(0)

	var anchor *Tile
	for x := 0; x < size; x++ {
		if anchor == nil {
			anchor = topLeft
		} else {
			anchor = anchor.AddDown(NewTile(0))
		}
		t := anchor
		for y := 0; y < size; y++ {
			t = t.AddRight(NewTile(0))
		}
	}

	return &Board{
		size:    size,
		topLeft: topLeft,
	}
}

func (b *Board) forEach(x, y int, tile *Tile, f func(int, int, *Tile) error) error {
	f(x, y, tile)
	if tile.right != nil {
		b.forEach(x+1, y, tile.right, f)
	}
	if tile.down != nil {
		b.forEach(x, y+1, tile.down, f)
	}
	return nil
}

// Update updates the board state.
func (b *Board) Update(input *Input) error {
	err := b.forEach(0, 0, b.topLeft, func(x, y int, t *Tile) error { return t.Update() })
	if err != nil {
		return fmt.Errorf("error updating tiles: %s", err)
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
			v := 0
			op := &ebiten.DrawImageOptions{}
			x := i*tileSize + (i+1)*tileMargin
			y := j*tileSize + (j+1)*tileMargin
			op.GeoM.Translate(float64(x), float64(y))
			r, g, b, a := colorToScale(tileBackgroundColor(v))
			op.ColorM.Scale(r, g, b, a)
			boardImage.DrawImage(tileImage, op)
		}
	}
	b.forEach(0, 0, b.topLeft, func(x, y int, t *Tile) error {
		t.Draw(x, y, boardImage)
		return nil
	})
}

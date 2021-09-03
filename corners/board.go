package corners

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Board represents the game board.
type Board struct {
	size    int
	topLeft *Tile
	offset  image.Point
}

// NewBoard generates a new Board with giving a size.
func NewBoard(size int) *Board {
	topLeft := NewTile(1)

	var anchor *Tile
	for x := 0; x < size; x++ {
		if anchor == nil {
			anchor = topLeft
		} else {
			anchor = anchor.AddDown(NewTile(1))
		}
		t := anchor
		for y := 0; y < size; y++ {
			t = t.AddRight(NewTile(1))
		}
	}

	return &Board{
		size:    size,
		topLeft: topLeft,
	}
}

func (b *Board) translate(mouse *image.Point) (int, int) {
	if mouse == nil {
		return -1, -1
	}
	p := mouse.Sub(b.offset)
	if p.X < 0 || p.Y < 0 {
		return -1, -1
	}
	x := p.X / tileSize
	y := p.Y / tileSize
	return x, y
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
	clickedX, clickedY := b.translate(input.LeftMouse())
	err := b.forEach(0, 0, b.topLeft, func(x, y int, t *Tile) error {
		return t.Update(&UpdateParams{
			clicked: x == clickedX && y == clickedY,
		})
	})
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

	// TODO fix this; it's hoping you have a square
	for j := 0; j < b.size; j++ {
		for i := 0; i < b.size; i++ {
			v := 0
			op := &ebiten.DrawImageOptions{}
			x := i*tileSize + (i+1)*tileMargin
			y := j*tileSize + (j+1)*tileMargin
			op.GeoM.Translate(float64(x), float64(y))
			r, g, b, a := colorToScale(tileBackgroundColor(v))
			op.ColorM.Scale(r, g, b, a)
			/*if j == 0 && i == 0 {
				fmt.Printf("drawing tile bg image: %+v\n", tileImage.Bounds())
			}*/
			boardImage.DrawImage(tileImage, op)
		}
	}
	b.forEach(0, 0, b.topLeft, func(x, y int, t *Tile) error {
		t.Draw(x, y, boardImage)
		return nil
	})
}

package corners

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Board represents the game board.
type Board struct {
	size     int
	tiles    [][]*Tile
	offset   image.Point
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
		tiles:    tiles,
		redbase:  redbase,
		bluebase: bluebase,
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
	x := p.X / (tileSize + tileMargin)
	y := p.Y / (tileSize + tileMargin)
	return x, y
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	for col := range b.tiles {
		for row := range b.tiles[col] {
			f(col, row, b.tiles[col][row])
		}
	}

	return nil
}

func boolptr(val bool) *bool {
	return &val
}

func (b *Board) selected() *Tile {
	var selected *Tile
	b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.selected {
			selected = t
		}
		return nil
	})
	return selected
}

func (b *Board) targeted() *Tile {
	var targeted *Tile
	b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.targeted {
			targeted = t
		}
		return nil
	})
	return targeted
}

// TODO make this try to follow a diagonal
func movementVector(from, to *Tile) (int, int) {
	if from.x < to.x {
		return 1, 0
	} else if from.x > to.x {
		return -1, 0
	}

	if from.y < to.y {
		return 0, 1
	} else if from.y > to.y {
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

		targetX, targetY := t.from.x+x, t.from.y+y

		// fmt.Printf("moving %d armies from %d,%d to %d,%d\n", t.armies, t.from.x, t.from.y, targetX, targetY)

		dest := b.tiles[targetX][targetY]

		if dest.team != 0 && dest.team != t.from.team {
			fmt.Println("calling attack!")
			t.from.attack(dest)
		}

		if dest.team == 0 || dest.team == t.from.team || dest.armies == 0 {
			t.armies = t.from.take(t.armies)
			dest.add(t.from, t.armies)
			t.from = dest
		} else {
			fmt.Println("grabbing low bar from armies")
			t.armies = int(math.Min(float64(t.from.armies), float64(t.armies)))
		}

		if t.from == t.to {
			return
		}

		fmt.Println("sleeping")
		<-ticker.C
	}
}

func (b *Board) transfer(source, dest *Tile) {
	armies := source.armies

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
func (b *Board) Update(input *Input) error {
	red, blue := 0, 0

	b.bluebase.resources = 0
	b.redbase.resources = 0

	b.forEach(0, 0, func(x, y int, t *Tile) error {
		if t.team == 1 {
			blue += t.resources
		} else if t.team == 2 {
			red += t.resources
		}
		return nil
	})

	// TODO lock this
	b.bluebase.resources = blue
	b.redbase.resources = red

	clickedX, clickedY := b.translate(input.LeftMouse())
	targetX, targetY := b.translate(input.RightMouse())

	selected := b.selected()
	targeted := b.targeted()

	// fmt.Printf("selected: %+v\n", selected)
	if selected != nil && targeted != nil && selected != targeted {
		b.transfer(selected, targeted)
	}

	err := b.forEach(0, 0, func(x, y int, t *Tile) error {
		params := UpdateParams{
			targeted: x == targetX && y == targetY,
		}
		if clickedX >= 0 && clickedY >= 0 {
			params.selected = boolptr(x == clickedX && y == clickedY && t.team != 0)
		}
		return t.Update(&params)
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
	b.forEach(0, 0, func(x, y int, t *Tile) error {
		t.Draw(x, y, boardImage)
		return nil
	})
}

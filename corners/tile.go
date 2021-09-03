package corners

import (
	"fmt"
	"image/color"
	"log"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"
)

var (
	fontSmall  font.Face
	fontNormal font.Face
	fontBig    font.Face
)

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	fontSmall, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	fontNormal, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    32,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	fontBig, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

// Tile represents a tile information including TileData and animation states.
type Tile struct {
	value int
	right *Tile
	left  *Tile
	down  *Tile
	up    *Tile
}

// NewTile creates a new Tile object.
func NewTile(value int) *Tile {
	return &Tile{
		value: value,
	}
}

func (t *Tile) AddRight(right *Tile) *Tile {
	t.right = right
	return right
}

// TODO fill in other left/right references once board is made
// - or don't, and let's make each board acycling so we can just walk any links a tile has?

func (t *Tile) AddDown(down *Tile) *Tile {
	t.down = down
	return down
}

// Update updates the tile's animation states.
func (t *Tile) Update() error {
	return nil
}

func colorToScale(clr color.Color) (float64, float64, float64, float64) {
	r, g, b, a := clr.RGBA()
	rf := float64(r) / 0xffff
	gf := float64(g) / 0xffff
	bf := float64(b) / 0xffff
	af := float64(a) / 0xffff
	// Convert to non-premultiplied alpha components.
	if 0 < af {
		rf /= af
		gf /= af
		bf /= af
	}
	return rf, gf, bf, af
}

const (
	tileSize   = 80
	tileMargin = 4
)

var (
	tileImage = ebiten.NewImage(tileSize, tileSize)
)

func init() {
	tileImage.Fill(color.White)
}

// Draw draws the current tile to the given boardImage.
func (t *Tile) Draw(x, y int, boardImage *ebiten.Image) {
	i, j := x, y

	op := &ebiten.DrawImageOptions{}
	x = i*tileSize + (i+1)*tileMargin
	y = j*tileSize + (j+1)*tileMargin
	op.GeoM.Translate(float64(x), float64(y))
	v := t.value
	r, g, b, a := colorToScale(tileBackgroundColor(v))
	op.ColorM.Scale(r, g, b, a)
	boardImage.DrawImage(tileImage, op)
	str := strconv.Itoa(v)

	f := fontBig
	if len(str) >= 3 {
		f = fontSmall
	} else if len(str) == 2 {
		f = fontNormal
	}

	bound, _ := font.BoundString(f, str)
	w := (bound.Max.X - bound.Min.X).Ceil()
	h := (bound.Max.Y - bound.Min.Y).Ceil()
	x = x + (tileSize-w)/2
	y = y + (tileSize-h)/2 + h
	fmt.Printf("drawing tile: str=%s x=%d y=%d\n", str, x, y)
	text.Draw(boardImage, str, f, x, y, tileColor(v))
}

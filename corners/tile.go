package corners

import (
	"image/color"
	"log"
	"strconv"
	"time"

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
	value     int
	team      int
	generator bool
	right     *Tile
	down      *Tile
	selected  bool
	targeted  bool
}

type TileParams struct {
	team      int
	generator bool
}

func (t *Tile) generate() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		t.value++
		<-ticker.C
	}
}

// NewTile creates a new Tile object.
func NewTile(params *TileParams) *Tile {
	t := &Tile{
		team:      params.team,
		value:     0,
		generator: params.generator,
	}
	if t.generator {
		go t.generate()
	}
	return t
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

type UpdateParams struct {
	selected *bool
	targeted bool
}

// Update updates the tile's animation states.
func (t *Tile) Update(params *UpdateParams) error {
	if params.selected != nil {
		t.selected = *params.selected
	}
	t.targeted = params.targeted
	return nil
}

// TODO rules that you can only select your own team squares

func (t *Tile) Target(target *Tile) {
	target.team = t.team
	target.value += t.value
	t.value = 0
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

func (t *Tile) bgColor() color.Color {
	if t.selected {
		return color.RGBA{0x00, 0x00, 0x00, 0xff}
	} else if t.targeted {
		return color.RGBA{0x00, 0x88, 0x00, 0xff}
	}

	switch t.team {
	case 1:
		return color.RGBA{0x00, 0x00, 0x88, 0xff}
	default:
		return color.NRGBA{0xee, 0xe4, 0xda, 0x59}
	}
}

// Draw draws the current tile to the given boardImage.
func (t *Tile) Draw(x, y int, boardImage *ebiten.Image) {
	i, j := x, y

	op := &ebiten.DrawImageOptions{}
	x = i*tileSize + (i+1)*tileMargin
	y = j*tileSize + (j+1)*tileMargin
	op.GeoM.Translate(float64(x), float64(y))
	r, g, b, a := colorToScale(t.bgColor())
	op.ColorM.Scale(r, g, b, a)
	v := t.value
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

	text.Draw(boardImage, str, f, x, y, tileColor(v))
}

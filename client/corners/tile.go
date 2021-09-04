package corners

import (
	"image/color"
	"log"
	"math"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"terrbear.io/corners/internal/rpc"

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
	armies int
	team   int
	x      int
	y      int

	tile *rpc.Tile
}

type TileParams struct {
	x         int
	y         int
	resources int
	team      int
	generator bool
}

// NewTile creates a new Tile object.
func NewTile(params *TileParams) *Tile {
	t := &Tile{
		x:      params.x,
		y:      params.y,
		team:   params.team,
		armies: 0,
	}
	return t
}

type UpdateParams struct {
	selected *bool
	targeted bool
}

func (t *Tile) adjacent(other *Tile) bool {
	return math.Abs(float64(t.x-other.x))+math.Abs(float64(t.y-other.y)) == 1
}

// TODO rules that you can only select your own team squares
// TODO make army transfers locked
// TODO wrap transfers in their own ob
// TODO can't reduce a tile's army <1
// TODO rename value to like 'army'

func (t *Tile) Target(target *Tile) {
	if t.adjacent(target) {
		target.team = t.team
		target.armies += t.armies
		t.armies = 0
	}
}

func colorToScale(clr color.Color) (float64, float64, float64, float64) {
	r, g, b, a := clr.RGBA()
	rf := float64(r) / 0xffff
	gf := float64(g) / 0xffff
	bf := float64(b) / 0xffff
	af := float64(a) / 0xffff
	// Convert to non-premultiplied alpha components.
	if af > 0 {
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

func (t *Tile) bgColor(params *TileDrawParams) color.Color {
	if params.selected {
		return color.RGBA{0x00, 0x00, 0x00, 0xff}
	} else if params.targeted {
		return color.RGBA{0x00, 0x88, 0x00, 0xff}
	}

	if t.tile != nil {
		alpha := uint8(0x33)

		if t.tile.Armies > 200 {
			alpha = 0xff
		} else if t.tile.Armies > 100 {
			alpha = 0xaa
		} else if t.tile.Armies > 50 {
			alpha = 0x88
		} else if t.tile.Armies > 10 {
			alpha = 0x66
		}

		switch t.tile.Team {
		case 1:
			return color.RGBA{0x00, 0x00, 0x88, alpha}
		case 2:
			return color.RGBA{0x88, 0x00, 0x00, alpha}
		}
	}

	return color.NRGBA{0xee, 0xe4, 0xda, 0x59}
}

type TileDrawParams struct {
	boardTeam int
	selected  bool
	targeted  bool
	team      int
}

// Draw draws the current tile to the given boardImage.
func (t *Tile) Draw(x, y int, boardImage *ebiten.Image, params *TileDrawParams) {
	i, j := x, y

	op := &ebiten.DrawImageOptions{}
	x = i*tileSize + (i+1)*tileMargin
	y = j*tileSize + (j+1)*tileMargin
	op.GeoM.Translate(float64(x), float64(y))
	r, g, b, a := colorToScale(t.bgColor(params))
	op.ColorM.Scale(r, g, b, a)
	v := 0
	if t.tile != nil {
		v = t.tile.Armies
	}
	boardImage.DrawImage(tileImage, op)
	str := strconv.Itoa(v)

	if t.tile != nil {
		if t.tile.Team == 0 && t.tile.Armies == 0 {
			return
		}
		if t.tile.Team != 0 && t.tile.Team != params.boardTeam {
			return
		}
	}

	f := fontBig
	if len(str) >= 3 {
		f = fontSmall
	} else if len(str) == 2 {
		f = fontNormal
	}

	// left ←
	// up ↑
	// right ↓

	//	str += "→"

	bound, _ := font.BoundString(f, str)
	w := (bound.Max.X - bound.Min.X).Ceil()
	h := (bound.Max.Y - bound.Min.Y).Ceil()
	x = x + (tileSize-w)/2
	y = y + (tileSize-h)/2 + h

	text.Draw(boardImage, str, f, x, y, tileColor(v))
}

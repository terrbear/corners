package corners

import (
	"image/color"
	"log"
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
	x int
	y int

	tile *rpc.Tile
}

type TileParams struct {
	x         int
	y         int
	resources int
}

// NewTile creates a new Tile object.
func NewTile(params *TileParams) *Tile {
	t := &Tile{
		x: params.x,
		y: params.y,
	}
	return t
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
		return color.RGBA{0x00, 0xaf, 0x00, 0x33}
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

		switch t.tile.PlayerID {
		case params.boardPlayerID:
			return color.RGBA{0x00, 0x00, 0x88, alpha} // blue
		default:
			c := getPlayerColor(t.tile.PlayerID)
			c.A = alpha
			return c
		}
	}

	return color.NRGBA{0xee, 0xe4, 0xda, 0x59}
}

type TileDrawParams struct {
	boardPlayerID rpc.PlayerID
	selected      bool
	targeted      bool
}

// Draw draws the current tile to the given boardImage.
func (t *Tile) Draw(xoffset, yoffset int, boardImage *ebiten.Image, params *TileDrawParams) {
	i, j := xoffset, yoffset

	op := &ebiten.DrawImageOptions{}
	x := float64(i*tileSize + (i+1)*tileMargin)
	y := float64(j*tileSize + (j+1)*tileMargin)

	if params.selected {
		scale := float64(tileSize+(tileMargin*2)) / tileSize
		x -= tileMargin
		y -= tileMargin
		op.GeoM.Scale(scale, scale)
		//op.GeoM.Translate(float64(x*2), float64(y*2))
		//op.GeoM.Fill(color.RGBA{0x00, 0x00, 0x00, 0xff})
	}
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
		if t.tile.PlayerID == rpc.NeutralPlayer && t.tile.Armies == 0 {
			return
		}
		if t.tile.PlayerID != rpc.NeutralPlayer && t.tile.PlayerID != params.boardPlayerID {
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
	x = x + float64((tileSize-w)/2)
	y = y + float64((tileSize-h)/2+h)

	c := color.RGBA{0xff, 0xff, 0xff, 0xff}
	if t.tile != nil && t.tile.PlayerID == rpc.NeutralPlayer {
		c = color.RGBA{0x00, 0x00, 0x00, 0xff}
	}
	text.Draw(boardImage, str, f, int(x), int(y), c)
}

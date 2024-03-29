package corners

import (
	"image/color"
	"strconv"

	"golang.org/x/image/font"
	"terrbear.io/corners/internal/rpc"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
)

// Tile represents a tile information including TileData and animation states.
type Tile struct {
	x int
	y int

	tile rpc.Tile
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
	tileSize   = 40
	tileMargin = 4
)

var (
	tileImage = ebiten.NewImage(tileSize, tileSize)
)

func init() {
	tileImage.Fill(color.White)
}

func (t *Tile) bgColor(params *TileDrawParams) color.Color {
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

type TileDrawParams struct {
	boardPlayerID rpc.PlayerID
	selected      bool
	targeted      bool
}

func (t *Tile) drawFog(i, j int, boardImage *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	x := float64(i*tileSize + (i+1)*tileMargin)
	y := float64(j*tileSize + (j+1)*tileMargin)
	op.GeoM.Translate(float64(x), float64(y))
	boardImage.DrawImage(tileImage, op)
	r, g, b, a := colorToScale(color.RGBA{0x0, 0x0, 0x0, 0x6a})
	op.ColorM.Scale(r, g, b, a)
	boardImage.DrawImage(tileImage, op)
}

func (t *Tile) drawDetailed(i, j int, boardImage *ebiten.Image, params *TileDrawParams) {
	op := &ebiten.DrawImageOptions{}
	x := float64(i*tileSize + (i+1)*tileMargin)
	y := float64(j*tileSize + (j+1)*tileMargin)
	op.GeoM.Translate(float64(x), float64(y))
	boardImage.DrawImage(tileImage, op)
	r, g, b, a := colorToScale(t.bgColor(params))
	op.ColorM.Scale(r, g, b, a)
	boardImage.DrawImage(tileImage, op)
	str := strconv.Itoa(t.tile.Armies)

	if t.tile.PlayerID == rpc.NeutralPlayer && t.tile.Armies == 0 {
		return
	}
	if t.tile.PlayerID != rpc.NeutralPlayer && t.tile.PlayerID != params.boardPlayerID {
		return
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
	if t.tile.PlayerID == rpc.NeutralPlayer {
		c = color.RGBA{0x00, 0x00, 0x00, 0xff}
	}
	text.Draw(boardImage, str, f, int(x), int(y), c)
}

func (t *Tile) Draw(xoffset, yoffset int, boardImage *ebiten.Image, params *TileDrawParams) {
	i, j := xoffset, yoffset

	if params.selected {
		// If the tile is selected, draw it green
		op := &ebiten.DrawImageOptions{}
		x := float64(i*tileSize + (i+1)*tileMargin - tileMargin)
		y := float64(j*tileSize + (j+1)*tileMargin - tileMargin)
		scale := float64(tileSize+(tileMargin*2)) / tileSize
		op.GeoM.Scale(scale, scale)

		op.GeoM.Translate(float64(x), float64(y))
		r, g, b, a := colorToScale(color.RGBA{0x00, 0xaf, 0x00, 0x33})
		op.ColorM.Scale(r, g, b, a)
		boardImage.DrawImage(tileImage, op)
	} else if t.tile.Generator {
		// If the tile is a generator, make it purple
		op := &ebiten.DrawImageOptions{}
		x := float64(i*tileSize + (i+1)*tileMargin - tileMargin)
		y := float64(j*tileSize + (j+1)*tileMargin - tileMargin)
		scale := float64(tileSize+(tileMargin*2)) / tileSize
		op.GeoM.Scale(scale, scale)

		op.GeoM.Translate(float64(x), float64(y))
		r, g, b, a := colorToScale(color.RGBA{0xa0, 0x45, 0xc5, 0x88})
		op.ColorM.Scale(r, g, b, a)
		boardImage.DrawImage(tileImage, op)
	}

	if t.tile.Visible {
		t.drawDetailed(i, j, boardImage, params)
	} else {
		t.drawFog(i, j, boardImage)
	}
}

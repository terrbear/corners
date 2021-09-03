package corners

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
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
	armies    int
	team      int
	generator bool
	x         int
	y         int
	lock      sync.Mutex
	resources int

	selected bool
	targeted bool
}

type TileParams struct {
	x         int
	y         int
	resources int
	team      int
	generator bool
}

func (t *Tile) generate() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		t.armies += (t.resources / 3) + 3
		<-ticker.C
	}
}

// NewTile creates a new Tile object.
func NewTile(params *TileParams) *Tile {
	t := &Tile{
		resources: params.resources,
		x:         params.x,
		y:         params.y,
		team:      params.team,
		armies:    0,
		generator: params.generator,
	}
	if t.generator {
		// to test early games faster
		t.armies = 20
		go t.generate()
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

// TODO rename this
func (t *Tile) add(other *Tile, armies int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.team = other.team
	fmt.Printf("adding %d armies to tile; current armies: %d\n", armies, t.armies)
	t.armies += armies
}

func (t *Tile) take(armies int) int {
	t.lock.Lock()
	defer t.lock.Unlock()
	if armies >= t.armies {
		available := t.armies - 1
		t.armies = 1
		return available
	}

	t.armies -= armies
	return armies
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

// Super simple risk rolling for now, returns values to take away from attacker and defender armies
func roll(attackers, defenders int) (int, int) {
	attacks := make([]int, attackers)
	defenses := make([]int, defenders)

	for i := 0; i < len(attacks); i++ {
		attacks[i] = rand.Intn(6) + 1
	}
	for i := 0; i < len(defenses); i++ {
		defenses[i] = rand.Intn(6) + 1
	}

	sort.SliceStable(attacks, func(i, j int) bool { return attacks[j] < attacks[i] })
	sort.SliceStable(defenses, func(i, j int) bool { return attacks[j] < attacks[i] })

	alosses, dlosses := 0, 0

	for i := 0; i < len(defenses); i++ {
		if defenses[i] >= attacks[i] {
			alosses++
		} else {
			dlosses++
		}
	}

	return alosses, dlosses
}

func (t *Tile) attack(defender *Tile) {
	defenders := 1
	if defender.armies > 1 {
		defenders = 2
	}

	attackers := 2
	if t.armies > 2 {
		attackers = 3
	}

	alosses, dlosses := roll(attackers, defenders)

	fmt.Printf("attacker loses %d armies, defender loses %d armies\n", alosses, dlosses)
	t.armies -= alosses
	defender.armies -= dlosses

	// TODO bottom that out at 0
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
	case 2:
		return color.RGBA{0x88, 0x00, 0x00, 0xff}
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
	v := t.armies
	boardImage.DrawImage(tileImage, op)
	str := strconv.Itoa(v)

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

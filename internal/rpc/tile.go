package rpc

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Tile represents a tile information including TileData and animation states.
type Tile struct {
	Armies    int `json:"armies"`
	Team      int `json:"team"`
	X         int `json:"x"`
	Y         int `json:"y"`
	resources int
	generator bool
	lock      sync.Mutex
	start     sync.Once
}

type TileParams struct {
	x         int
	y         int
	resources int
	team      int
	generator bool
}

func (t *Tile) generate() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		t.Armies += (t.resources / 6) + 3
		<-ticker.C
	}
}

func (t *Tile) Start() {
	if t.generator {
		t.start.Do(func() {
			go t.generate()
		})
	}
}

// NewTile creates a new Tile object.
func NewTile(params *TileParams) *Tile {
	t := &Tile{
		resources: params.resources,
		X:         params.x,
		Y:         params.y,
		Team:      params.team,
		Armies:    0,
		generator: params.generator,
	}
	if t.generator {
		// to test early games faster
		t.Armies = 10
	}
	return t
}

// TODO rename this
func (t *Tile) add(other *Tile, armies int) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.Team = other.Team
	fmt.Printf("adding %d armies to tile; current armies: %d\n", armies, t.Armies)
	t.Armies += armies
}

func (t *Tile) take(armies int) int {
	t.lock.Lock()
	defer t.lock.Unlock()
	if armies >= t.Armies {
		available := t.Armies - 1
		t.Armies = 1
		return available
	}

	t.Armies -= armies
	return armies
}

// TODO rules that you can only select your own team squares
// TODO make army transfers locked
// TODO wrap transfers in their own ob
// TODO can't reduce a tile's army <1
// TODO rename value to like 'army'

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
	t.lock.Lock()
	defender.lock.Lock()

	defer t.lock.Unlock()
	defer defender.lock.Unlock()

	if t.Armies <= 1 {
		return
	}

	defenders := 1
	if defender.Armies > 1 {
		defenders = 2
	}

	attackers := 2
	if t.Armies > 2 {
		attackers = 3
	}

	alosses, dlosses := roll(attackers, defenders)

	fmt.Printf("attacker loses %d armies, defender loses %d armies\n", alosses, dlosses)
	t.Armies -= alosses
	defender.Armies -= dlosses

	if defender.Armies < 0 {
		defender.Armies = 0
	}

	if t.Armies < 0 {
		t.Armies = 1
	}
}

package corners

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"terrbear.io/corners/internal/rpc"
)
import "github.com/beefsack/go-astar"

type PathableTile struct {
	board    *Board
	tile     *Tile
	playerID rpc.PlayerID
}

func (p PathableTile) PathNeighbors() []astar.Pather {
	var pathables []astar.Pather

	tiles := p.board.TileNeighbours(p.tile)
	for i := range p.board.TileNeighbours(p.tile) {
		pathables = append(pathables, PathableTile{
			tile: tiles[i],
			board: p.board,
			playerID: tiles[i].PlayerID,
		})
	}

	return pathables
}

func (p PathableTile) PathNeighborCost(to astar.Pather) float64 {
	pathable := to.(PathableTile)
	tile := pathable.tile
	cost := float64(0)

	if tile.PlayerID != p.tile.PlayerID {
	//	cost = float64(tile.Armies)
		cost = float64(tile.Armies)
	}

	//if cost == 1 {
	//	cost = 0
	//}

	//log.Debugf(
	//	"(%d, %d) '%s' vs '%s', '%s' vs '%s' - cost is %f",
	//	pathable.tile.X,
	//	pathable.tile.Y,
	//	p.tile.PlayerID,
	//	tile.PlayerID,
	//	p.playerID,
	//	pathable.playerID,
	//	cost,
	//)

	//if cost > 0 {
	//	panic('cost should not be greater than 0')
	//}

	return cost
}

func (p PathableTile) PathEstimatedCost(to astar.Pather) float64 {
	return p.tile.Distance(to.(PathableTile).tile)
}

func printPath(path []astar.Pather) {
	var str []string

	for _, pather := range path {

		str = append(str, fmt.Sprintf(
			"(%d, %d)",
			pather.(PathableTile).tile.X,
			pather.(PathableTile).tile.Y,
		))
	}

	log.Debug(strings.Join(str[:], " -> "))
}

func FindPath(from, to *Tile, board *Board) (rpc.Point, bool) {
	point := rpc.Point{0, 0}

	path, _, found := astar.Path(
		PathableTile{tile: to, board: board, playerID: from.PlayerID},
		PathableTile{tile: from, board: board, playerID: from.PlayerID},
	)

	if found {
		log.Debug("Found path!")
		printPath(path)
		tile := path[1].(PathableTile).tile

		point.X = tile.X
		point.Y = tile.Y
	} else {
		log.Debug("Failed to find path...")
	}

	return point, found
}

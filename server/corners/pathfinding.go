package corners

import (
	log "github.com/sirupsen/logrus"
	"terrbear.io/corners/internal/rpc"
)
import "github.com/beefsack/go-astar"

type PathableTile struct {
	board    *Board
	tile     *Tile
}

func (p PathableTile) PathNeighbors() []astar.Pather {
	var pathables []astar.Pather

	for _, tile := range p.board.TileNeighbours(p.tile) {
		pathables = append(pathables, PathableTile{tile: tile, board: p.board})
	}

	return pathables
}

func (p PathableTile) PathNeighborCost(to astar.Pather) float64 {
	//tile := to.(PathableTile).tile

	//if tile.PlayerID != p.playerID {
	//	cost := float64(tile.Armies)
	//
	//	log.Debugf(
	//		"'%s' vs '%s' vs '%s' - tile cost is %f",
	//		//"Different player ids: %s vs %s - tile cost is %f",
	//		p.tile.PlayerID,
	//		tile.PlayerID,
	//		p.playerID,
	//		cost,
	//	)
	//
	//	return cost
	//} else {
	//	log.Debugf(
	//		"'%s' vs '%s' vs '%s'",
	//		//"Different player ids: %s vs %s - tile cost is %f",
	//		p.tile.PlayerID,
	//		tile.PlayerID,
	//		p.playerID,
	//	)
	//}

	return 0
}

func (p PathableTile) PathEstimatedCost(to astar.Pather) float64 {
	return p.tile.Distance(to.(PathableTile).tile)
}

func FindPath(from, to *Tile, board *Board) (rpc.Point, bool) {
	point := rpc.Point{0, 0}

	path, _, found := astar.Path(
		PathableTile{tile: from, board: board},
		PathableTile{tile: to, board: board},
	)

	if found {
		log.Debug("Found path!")
		tile := path[1].(PathableTile).tile

		point.X = tile.X
		point.Y = tile.Y
	} else {
		log.Debug("Failed to find path...")
	}

	return point, found
}

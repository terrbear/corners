package rpc

// Board represents the game board.
type Board struct {
	Tiles [][]*Tile `json:"tiles"`
}

type PlayerID string

const NeutralPlayer PlayerID = "neutral"

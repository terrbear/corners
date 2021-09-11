package rpc

import (
	"encoding/json"
)

// Board represents the game board.
type Board struct {
	Tiles  [][]Tile  `json:"tiles"`
	Winner *PlayerID `json:"winner,omitempty"`
}

type PlayerID string

const NeutralPlayer PlayerID = "neutral"

func SerializeBoard(b *Board) ([]byte, error) {
	return json.Marshal(b)
}

func DeserializeBoard(b []byte) (*Board, error) {
	var board Board
	if err := json.Unmarshal(b, &board); err != nil {
		return nil, err
	}
	return &board, nil
}

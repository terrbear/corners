package rpc

import "github.com/kelindar/binary"

// Board represents the game board.
type Board struct {
	Tiles [][]Tile `json:"tiles"`
}

type PlayerID string

const NeutralPlayer PlayerID = "neutral"

func SerializeBoard(b *Board) ([]byte, error) {
	return binary.Marshal(b)
}

func DeserializeBoard(b []byte) (*Board, error) {
	var board Board
	if err := binary.Unmarshal(b, &board); err != nil {
		return nil, err
	}
	return &board, nil
}

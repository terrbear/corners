package rpc

// Board represents the game board.
type Board struct {
	Tiles  [][]Tile  `json:"tiles"`
	Winner *PlayerID `json:"winner,omitempty"`
}

type PlayerID string

const NeutralPlayer PlayerID = "neutral"

func SerializeBoard(b *Board) ([]byte, error) {
	return serialize(b)
}

func DeserializeBoard(b []byte) (*Board, error) {
	var board Board
	if err := deserialize(b, &board); err != nil {
		return nil, err
	}
	return &board, nil
}

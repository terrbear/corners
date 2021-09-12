package rpc

// Board represents the game board.
type Board struct {
	Size   int       `json:"size"`
	Tiles  []Tile    `json:"tiles"`
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

func (b *Board) Diff(other *Board) *Board {
	if other == nil {
		return b
	}

	diff := Board{
		Winner: b.Winner,
	}

	for i, t := range b.Tiles {
		hasChanged := other.Tiles[i] != t
		visibilityChanged := t.Visible != other.Tiles[i].Visible

		if (t.Visible && hasChanged) || visibilityChanged {
			diff.Tiles = append(diff.Tiles, t)
		}
	}

	return &diff
}

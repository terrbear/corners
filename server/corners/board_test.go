package corners

import (
	"testing"

	"terrbear.io/corners/internal/rpc"
)

func TestCanMakeBoard(t *testing.T) {
	NewBoard("test/og", []rpc.PlayerID{"hi"})
}

func TestWinCondition(t *testing.T) {
	winner := rpc.PlayerID("hi")
	b := NewBoard("test/og", []rpc.PlayerID{winner})
	if b.Winner != nil {
		t.Error("new board shouldn't have a winner")
	}
	maxIdx := len(b.Tiles) - 1
	// This is tightly coupled to the OG map, which has generators at each corner.
	b.Tiles[0][0].PlayerID = winner
	b.Tiles[maxIdx][0].PlayerID = winner
	b.Tiles[0][maxIdx].PlayerID = winner
	b.Tiles[maxIdx][maxIdx].PlayerID = winner

	b.tick()

	if *b.Winner != winner {
		t.Error("board with winner taking over all generators should have a winner")
	}
}

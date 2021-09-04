package corners

import (
	"image/color"

	"terrbear.io/corners/internal/rpc"
)

var (
	backgroundColor = color.RGBA{0xfa, 0xf8, 0xef, 0xff}
	frameColor      = color.RGBA{0xbb, 0xad, 0xa0, 0xff}
)

func colorToScale(clr color.Color) (float64, float64, float64, float64) {
	r, g, b, a := clr.RGBA()
	rf := float64(r) / 0xffff
	gf := float64(g) / 0xffff
	bf := float64(b) / 0xffff
	af := float64(a) / 0xffff
	// Convert to non-premultiplied alpha components.
	if af > 0 {
		rf /= af
		gf /= af
		bf /= af
	}
	return rf, gf, bf, af
}

var playerMap = map[rpc.PlayerID]color.RGBA{
	rpc.NeutralPlayer: {0xee, 0xe4, 0xda, 0x59},
}

var playerColors = []color.RGBA{
	{0x44, 0x44, 0x00, 0x00}, // ??
	{0x88, 0x00, 0x00, 0x00}, // red
	{0x00, 0x44, 0x44, 0x00}, // ??
}

func getPlayerColor(playerID rpc.PlayerID) color.RGBA {
	if color, ok := playerMap[playerID]; ok {
		return color
	}
	playerMap[playerID] = playerColors[len(playerMap)%len(playerColors)]
	return playerMap[playerID]
}

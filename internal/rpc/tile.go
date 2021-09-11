package rpc

// Tile represents a tile information including TileData and animation states.
type Tile struct {
	Armies    int      `json:"armies"`
	PlayerID  PlayerID `json:"playerID"`
	X         int      `json:"x"`
	Y         int      `json:"y"`
	Generator bool     `json:"generator"`
	Visible   bool     `json:"visible"`
}

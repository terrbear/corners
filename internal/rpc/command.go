package rpc

type Command struct {
	SelectedX int `json:"selectedX"`
	SelectedY int `json:"selectedY"`
	TargetX   int `json:"targetX"`
	TargetY   int `json:"targetY"`
}

package rpc

import "github.com/kelindar/binary"

type Command struct {
	SelectedX int `json:"selectedX"`
	SelectedY int `json:"selectedY"`
	TargetX   int `json:"targetX"`
	TargetY   int `json:"targetY"`
}

func SerializeCommand(b *Command) ([]byte, error) {
	return binary.Marshal(b)
}

func DeserializeCommand(b []byte) (*Command, error) {
	var command Command
	if err := binary.Unmarshal(b, &command); err != nil {
		return nil, err
	}
	return &command, nil
}

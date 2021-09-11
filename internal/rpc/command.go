package rpc

import "encoding/json"

type Command struct {
	SelectedX int `json:"selectedX"`
	SelectedY int `json:"selectedY"`
	TargetX   int `json:"targetX"`
	TargetY   int `json:"targetY"`
}

func SerializeCommand(b *Command) ([]byte, error) {
	return json.Marshal(b)
}

func DeserializeCommand(b []byte) (*Command, error) {
	var command Command
	if err := json.Unmarshal(b, &command); err != nil {
		return nil, err
	}
	return &command, nil
}

package rpc

type Command struct {
	SelectedX int `json:"selectedX"`
	SelectedY int `json:"selectedY"`
	TargetX   int `json:"targetX"`
	TargetY   int `json:"targetY"`
}

func SerializeCommand(b *Command) ([]byte, error) {
	return serialize(b)
}

func DeserializeCommand(b []byte) (*Command, error) {
	var command Command
	if err := deserialize(b, &command); err != nil {
		return nil, err
	}
	return &command, nil
}

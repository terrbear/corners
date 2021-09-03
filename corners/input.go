package corners

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

type mouseState int

const (
	mouseStateNone mouseState = iota
	mouseStatePressing
	mouseStateSettled
)

// Input represents the current key states.
type Input struct {
	mouseState mouseState
	mouseX     int
	mouseY     int
}

// NewInput generates a new Input object.
func NewInput() *Input {
	return &Input{}
}

// Update updates the current input states.
func (i *Input) Update() {
	switch i.mouseState {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.mouseX = x
			i.mouseY = y
			i.mouseState = mouseStatePressing
			fmt.Println("left mouse click: ", x, y)
		}
	case mouseStatePressing:
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			i.mouseState = mouseStateSettled
		}
	case mouseStateSettled:
		i.mouseState = mouseStateNone
	}
}

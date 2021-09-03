package corners

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type mouseState int

const (
	mouseStateNone mouseState = iota
	mouseStateDown
	mouseStateUp
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

func (i *Input) LeftMouse() *image.Point {
	if i.mouseState == mouseStateDown {
		return &image.Point{i.mouseX, i.mouseY}
	}
	return nil
}

// Update updates the current input states.
func (i *Input) Update() {
	switch i.mouseState {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.mouseX = x
			i.mouseY = y
			i.mouseState = mouseStateDown
			fmt.Println("left mouse click: ", x, y)
		}
	case mouseStateDown:
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			i.mouseState = mouseStateUp
		}
	case mouseStateUp:
		i.mouseState = mouseStateNone
	}
}

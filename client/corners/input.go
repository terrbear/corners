package corners

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	log "github.com/sirupsen/logrus"
)

type mouseState int

const (
	mouseStateNone mouseState = iota
	mouseStateSettled
	mouseStateDown
	mouseStateUp
)

type MouseInput struct {
	state mouseState
	x     int
	y     int
}

// Input represents the current key states.
type Input struct {
	left     MouseInput
	right    MouseInput
	spacebar bool
}

// NewInput generates a new Input object.
func NewInput() *Input {
	return &Input{}
}

func (i *Input) LeftMouse() *image.Point {
	if i.left.state == mouseStateDown {
		return &image.Point{i.left.x, i.left.y}
	}
	return nil
}

func (i *Input) RightMouse() *image.Point {
	if i.right.state == mouseStateDown {
		return &image.Point{i.right.x, i.right.y}
	}
	return nil
}

// Update updates the current input states.
func (i *Input) Update() {
	switch i.left.state {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			i.left.x = x
			i.left.y = y
			i.left.state = mouseStateDown
			log.Debug("left mouse click: ", x, y)
		}
	case mouseStateDown, mouseStateSettled:
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			i.left.state = mouseStateUp
		} else {
			i.left.state = mouseStateSettled
		}
	case mouseStateUp:
		i.left.state = mouseStateNone
	}

	switch i.right.state {
	case mouseStateNone:
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			x, y := ebiten.CursorPosition()
			i.right.x = x
			i.right.y = y
			i.right.state = mouseStateDown
			log.Debug("right mouse click: ", x, y)
		}
	case mouseStateDown, mouseStateSettled:
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			i.right.state = mouseStateUp
		} else {
			i.right.state = mouseStateSettled
		}
	case mouseStateUp:
		i.right.state = mouseStateNone
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		i.spacebar = true
	}
}

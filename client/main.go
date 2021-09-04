package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"terrbear.io/corners/client/corners"
)

func main() {
	game, err := corners.NewGame()
	if err != nil {
		log.Fatal(err)
	}
	ebiten.SetWindowSize(corners.ScreenWidth, corners.ScreenHeight)
	ebiten.SetWindowTitle("CORNERS!!")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

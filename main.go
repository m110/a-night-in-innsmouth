package main

import (
	"flag"
	"log"

	"github.com/m110/secrets/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	quickFlag := flag.Bool("quick", false, "quick mode")
	flag.Parse()
	_ = quickFlag

	config := game.Config{
		Quick:        true,
		ScreenWidth:  1300,
		ScreenHeight: 768,
	}

	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)

	err := ebiten.RunGame(game.NewGame(config))
	if err != nil {
		log.Fatal(err)
	}
}

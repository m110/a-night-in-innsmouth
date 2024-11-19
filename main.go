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
		Quick: true,
	}

	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowSize(1920, 1080)

	err := ebiten.RunGame(game.NewGame(config))
	if err != nil {
		log.Fatal(err)
	}
}

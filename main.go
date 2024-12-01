package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/engine"
	"github.com/m110/secrets/game"
)

func main() {
	quickFlag := flag.Bool("quick", false, "quick mode")
	flag.Parse()

	config := game.Config{
		Quick: *quickFlag,
	}

	ebiten.SetVsyncEnabled(true)
	w, h := ebiten.Monitor().Size()
	ebiten.SetWindowSize(engine.IntPercent(w, 0.8), engine.IntPercent(h, 0.8))

	err := ebiten.RunGame(game.NewGame(config))
	if err != nil {
		log.Fatal(err)
	}
}

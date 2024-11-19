package mobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/m110/secrets/game"
)

func init() {
	mobile.SetGame(game.NewGame(game.Config{
		Quick: true,
	}))
}

func Dummy() {}

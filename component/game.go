package component

import (
	"github.com/m110/secrets/engine"
	"github.com/yohamta/donburi"
)

type GameData struct {
	Settings Settings
}

type Settings struct {
	ScreenWidth  int
	ScreenHeight int
}

var Game = donburi.NewComponentType[GameData]()

func MustFindGame(w donburi.World) *GameData {
	return engine.MustFindComponent[GameData](w, Game)
}

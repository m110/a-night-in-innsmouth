package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"

	"github.com/m110/secrets/engine"
)

type GameData struct {
	Story    *domain.Story
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

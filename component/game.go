package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

type GameData struct {
	Story *domain.Story

	// Can be saved to go "back" to the previous level, keeping the player's position
	PreviousLevel *PreviousLevel

	Settings Settings
}

type PreviousLevel struct {
	Name              string
	CharacterPosition *domain.CharacterPosition
}

type Settings struct {
	ScreenWidth  int
	ScreenHeight int
}

var Game = donburi.NewComponentType[GameData]()

func MustFindGame(w donburi.World) *GameData {
	return engine.MustFindComponent[GameData](w, Game)
}

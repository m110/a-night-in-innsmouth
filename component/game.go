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

	Dimensions Dimensions
}

type PreviousLevel struct {
	Name              string
	CharacterPosition *domain.CharacterPosition
}

type Dimensions struct {
	Updated bool

	ScreenWidth  int
	ScreenHeight int

	InventoryWidth int

	DialogWidth            int
	DialogLogHeight        int
	DialogOptionsHeight    int
	DialogOptionsRowHeight int
}

var Game = donburi.NewComponentType[GameData]()

func MustFindGame(w donburi.World) *GameData {
	return engine.MustFindComponent[GameData](w, Game)
}

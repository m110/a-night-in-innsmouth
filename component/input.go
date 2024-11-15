package component

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type InputData struct {
	Disabled bool

	MoveRightKeys []ebiten.Key
	MoveLeftKeys  []ebiten.Key
	ActionKeys    []ebiten.Key

	MoveSpeed float64
	ShootKey  ebiten.Key
}

var Input = donburi.NewComponentType[InputData]()

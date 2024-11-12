package component

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type InputData struct {
	Disabled bool

	MoveRightKey ebiten.Key
	MoveLeftKey  ebiten.Key
	ActionKey    ebiten.Key

	MoveSpeed float64

	ShootKey ebiten.Key
}

var Input = donburi.NewComponentType[InputData]()

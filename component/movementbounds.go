package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type MovementBoundsData struct {
	Range engine.FloatRange
}

var MovementBounds = donburi.NewComponentType[MovementBoundsData]()

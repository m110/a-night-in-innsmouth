package component

import "github.com/yohamta/donburi"

type MovementBoundsData struct {
	MinX float64
	MaxX float64
}

var MovementBounds = donburi.NewComponentType[MovementBoundsData]()

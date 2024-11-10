package component

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/engine"
)

type BoundsData struct {
	Width  float64
	Height float64
}

func (d BoundsData) Rect(e *donburi.Entry) engine.Rect {
	pos := transform.WorldPosition(e)
	return engine.NewRect(
		pos.X,
		pos.Y,
		d.Width,
		d.Height,
	)
}

var Bounds = donburi.NewComponentType[BoundsData]()

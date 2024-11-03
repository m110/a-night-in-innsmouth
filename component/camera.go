package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type CameraData struct {
	Zoom engine.FloatRange
}

var Camera = donburi.NewComponentType[CameraData]()

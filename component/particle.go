package component

import (
	"image/color"

	"github.com/yohamta/donburi"
)

type ParticleData struct {
	Life    int
	MaxLife int
	Size    float64
	Color   color.RGBA
}

var Particle = donburi.NewComponentType[ParticleData]()

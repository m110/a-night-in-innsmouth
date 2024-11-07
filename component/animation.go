package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type AnimationData struct {
	Active bool
	Timer  *engine.Timer
	Update func(e *donburi.Entry)
}

func (a *AnimationData) Stop() {
	a.Active = false
}

func (a *AnimationData) Start() {
	a.Active = true
	a.Timer.Reset()
}

var Animation = donburi.NewComponentType[AnimationData]()

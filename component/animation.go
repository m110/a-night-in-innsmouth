package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type AnimationData struct {
	Active  bool
	Timer   *engine.Timer
	Update  func(e *donburi.Entry)
	OnStart func(e *donburi.Entry)
	OnStop  func(e *donburi.Entry)
}

func (a *AnimationData) Stop(e *donburi.Entry) {
	a.Active = false

	if a.OnStop != nil {
		a.OnStop(e)
	}
}

func (a *AnimationData) Start(e *donburi.Entry) {
	if a.Active {
		return
	}

	a.Active = true
	a.Timer.Reset()

	if a.OnStart != nil {
		a.OnStart(e)
	}
}

var Animation = donburi.NewComponentType[AnimationData]()

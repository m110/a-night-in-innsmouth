package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type AnimatorData struct {
	Animations map[string]*Animation
}

func (a *AnimatorData) SetAnimation(name string, anim *Animation) {
	if a.Animations == nil {
		a.Animations = make(map[string]*Animation)
	}
	a.Animations[name] = anim
}

func (a *AnimatorData) Stop(name string, e *donburi.Entry) {
	anim, ok := a.Animations[name]
	if !ok {
		panic("animation not found: " + name)
	}

	anim.Stop(e)
}

func (a *AnimatorData) Start(name string, e *donburi.Entry) {
	anim, ok := a.Animations[name]
	if !ok {
		panic("animation not found: " + name)
	}

	anim.Start(e)
}

type Animation struct {
	Active         bool
	Timer          *engine.Timer
	Update         func(e *donburi.Entry, a *Animation)
	OnStart        func(e *donburi.Entry)
	OnStartOneShot []func(e *donburi.Entry)
	OnStop         func(e *donburi.Entry)
	OnStopOneShot  []func(e *donburi.Entry)
}

func (a *Animation) Stop(e *donburi.Entry) {
	a.Active = false

	if a.OnStop != nil {
		a.OnStop(e)
	}

	for _, f := range a.OnStopOneShot {
		f(e)
	}

	a.OnStopOneShot = nil
}

func (a *Animation) Start(e *donburi.Entry) {
	if a.Active {
		return
	}

	a.Active = true
	a.Timer.Reset()

	if a.OnStart != nil {
		a.OnStart(e)
	}

	for _, f := range a.OnStartOneShot {
		f(e)
	}

	a.OnStartOneShot = nil
}

var Animator = donburi.NewComponentType[AnimatorData]()

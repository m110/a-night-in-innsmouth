package component

import "github.com/yohamta/donburi"

type ActiveData struct {
	Active bool
}

func (a *ActiveData) Toggle() {
	a.Active = !a.Active
}

var Active = donburi.NewComponentType[ActiveData]()

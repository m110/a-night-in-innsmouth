package component

import "github.com/yohamta/donburi"

type ActiveData struct {
	Active bool
}

var Active = donburi.NewComponentType[ActiveData]()

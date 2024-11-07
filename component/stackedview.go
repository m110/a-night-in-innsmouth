package component

import "github.com/yohamta/donburi"

type StackedViewData struct {
	CurrentY float64
	Scrolled bool
}

var StackedView = donburi.NewComponentType[StackedViewData]()

package component

import "github.com/yohamta/donburi"

type POIData struct {
	Passage string
}

var POI = donburi.NewComponentType[POIData]()

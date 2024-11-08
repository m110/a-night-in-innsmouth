package component

import "github.com/yohamta/donburi"

type TagData struct {
	// Used for debugging
	Tag string
}

var Tag = donburi.NewComponentType[TagData]()

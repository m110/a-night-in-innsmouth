package component

import "github.com/yohamta/donburi"

type DialogOptionData struct {
	Index int
}

var DialogOption = donburi.NewComponentType[DialogOptionData]()

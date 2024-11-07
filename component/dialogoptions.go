package component

import "github.com/yohamta/donburi"

type DialogOptionData struct {
	Index int
	Lines int
}

var DialogOption = donburi.NewComponentType[DialogOptionData]()

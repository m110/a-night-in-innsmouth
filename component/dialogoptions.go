package component

import "github.com/yohamta/donburi"

type DialogOptionData struct {
	Index int
	Lines int
}

func (d DialogOptionData) Order() int {
	return d.Index
}

var DialogOption = donburi.NewComponentType[DialogOptionData]()

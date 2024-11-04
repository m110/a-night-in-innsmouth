package component

import "github.com/yohamta/donburi"

type DialogData struct {
	Passage      *Passage
	ActiveOption int
}

var Dialog = donburi.NewComponentType[DialogData]()

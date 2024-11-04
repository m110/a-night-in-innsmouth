package component

import "github.com/yohamta/donburi"

type DialogOption struct {
	Effect func()
	Text   string
}

type DialogData struct {
	Text         string
	Options      []DialogOption
	ActiveOption int
	JustChanged  bool
}

var Dialog = donburi.NewComponentType[DialogData]()

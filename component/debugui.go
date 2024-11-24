package component

import (
	"github.com/ebitenui/ebitenui"
	"github.com/yohamta/donburi"
)

type DebugUIData struct {
	UI *ebitenui.UI
}

var DebugUI = donburi.NewComponentType[DebugUIData]()

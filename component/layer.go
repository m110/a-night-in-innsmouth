package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/definitions"
)

type LayerData struct {
	Layer definitions.LayerID
}

var Layer = donburi.NewComponentType[LayerData]()

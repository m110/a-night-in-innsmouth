package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"
)

type LayerData struct {
	Layer domain.LayerID
}

var Layer = donburi.NewComponentType[LayerData]()

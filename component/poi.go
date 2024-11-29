package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"
)

type POIData struct {
	POI          domain.POI
	ParentObject *donburi.Entry
}

var POI = donburi.NewComponentType[POIData]()

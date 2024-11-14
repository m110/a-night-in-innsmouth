package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

type DetectPOI struct {
	poiQuery       *donburi.Query
	characterQuery *donburi.Query
}

func NewDetectPOI() *DetectPOI {
	return &DetectPOI{
		poiQuery:       donburi.NewQuery(filter.Contains(component.POIImage)),
		characterQuery: donburi.NewQuery(filter.Contains(component.Character)),
	}
}

func (d *DetectPOI) Update(w donburi.World) {
	character, ok := d.characterQuery.First(w)
	if !ok {
		return
	}

	characterPos := transform.WorldPosition(character)

	d.poiQuery.Each(w, func(poi *donburi.Entry) {
		poiPos := transform.WorldPosition(poi)
		distance := characterPos.Distance(poiPos)
		component.Sprite.Get(poi).ColorBlendOverride.Value = distanceToBlendValue(distance)
	})
}

var poiVisibleDistance = engine.FloatRange{Min: 400, Max: 1000}

func distanceToBlendValue(currentDist float64) float64 {
	if currentDist <= poiVisibleDistance.Min {
		return 1
	}

	if currentDist >= poiVisibleDistance.Max {
		return 0
	}

	return 1 - (currentDist-poiVisibleDistance.Min)/(poiVisibleDistance.Max-poiVisibleDistance.Min)
}

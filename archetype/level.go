package archetype

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
)

func NewLevel(w donburi.World, level domain.Level) *donburi.Entry {
	entry := NewTagged(w, "Level").
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: level.Background,
		}).
		Entry()

	for _, poi := range level.POIs {
		NewPOI(entry, poi)
	}

	return entry
}

package archetype

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
)

func NewLevel(w donburi.World, levelName string) *donburi.Entry {
	level := assets.Levels[levelName]

	entry := NewTagged(w, "Level").
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: level.Background,
		}).
		With(component.Level).
		Entry()

	for _, poi := range level.POIs {
		NewPOI(entry, poi)
	}

	return entry
}

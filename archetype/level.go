package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/component"
)

func NewLevel(w donburi.World, image *ebiten.Image) *donburi.Entry {
	level := NewTagged(w, "Level").
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: image,
		}).
		Entry()

	return level
}

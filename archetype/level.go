package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/component"
)

func NewLevel(w donburi.World, image *ebiten.Image) *donburi.Entry {
	level := NewTagged(w, "Level").
		WithPosition(math.Vec2{
			X: 50,
			Y: 50,
		}).
		WithScale(math.Vec2{
			X: 0.4,
			Y: 0.4,
		}).
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: image,
		}).
		Entry()

	return level
}

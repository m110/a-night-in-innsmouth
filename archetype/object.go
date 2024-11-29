package archetype

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
)

func NewObject(
	parent *donburi.Entry,
	obj domain.Object,
) *donburi.Entry {
	w := parent.World

	layer := domain.SpriteLayerObjects
	if obj.Layer != 0 {
		layer = obj.Layer
	}

	entry := NewTagged(w, "Object").
		WithParent(parent).
		WithPosition(obj.Position).
		WithScale(obj.Scale).
		WithLayer(layer).
		WithSprite(component.SpriteData{
			Image: obj.Image,
		}).
		WithSpriteBounds().
		Entry()

	return entry
}

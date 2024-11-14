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

	entry := NewTagged(w, "POI").
		WithParent(parent).
		WithPosition(obj.Position).
		WithLayer(obj.Layer).
		WithSprite(component.SpriteData{
			Image: obj.Image,
		}).
		WithSpriteBounds().
		Entry()

	return entry
}

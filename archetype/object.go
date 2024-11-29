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

	entry := NewTagged(w, "Object").
		WithParent(parent).
		WithPosition(obj.Position).
		WithScale(obj.Scale).
		WithLayer(obj.Layer).
		WithSprite(component.SpriteData{
			Image: obj.Image,
		}).
		WithSpriteBounds().
		Entry()

	return entry
}

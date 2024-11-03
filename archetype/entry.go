package archetype

import (
	"github.com/yohamta/donburi"
	donburicomponent "github.com/yohamta/donburi/component"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/component"
)

type EntryBuilder struct {
	entry *donburi.Entry
}

// New creates a new entry with Transform.
func New(w donburi.World) EntryBuilder {
	return EntryBuilder{
		entry: w.Entry(w.Create(transform.Transform)),
	}
}

func (b EntryBuilder) With(c donburicomponent.IComponentType) EntryBuilder {
	if !b.entry.HasComponent(c) {
		b.entry.AddComponent(c)
	}

	return b
}

func (b EntryBuilder) WithPosition(pos math.Vec2) EntryBuilder {
	transform.Transform.Get(b.entry).LocalPosition = pos
	return b
}

func (b EntryBuilder) WithScale(scale math.Vec2) EntryBuilder {
	transform.Transform.Get(b.entry).LocalScale = scale
	return b
}

func (b EntryBuilder) WithParent(parent *donburi.Entry) EntryBuilder {
	transform.AppendChild(parent, b.entry, false)
	return b
}

func (b EntryBuilder) WithSprite(sprite component.SpriteData) EntryBuilder {
	if !b.entry.HasComponent(component.Layer) {
		b.With(component.Layer)
	}
	b.With(component.Sprite)
	component.Sprite.SetValue(b.entry, sprite)
	return b
}

func (b EntryBuilder) WithLayer(layer component.LayerID) EntryBuilder {
	b.With(component.Layer)
	component.Layer.Get(b.entry).Layer = layer
	return b
}

func (b EntryBuilder) WithText(text component.TextData) EntryBuilder {
	if !b.entry.HasComponent(component.Layer) {
		b.With(component.Layer)
	}
	b.With(component.Text)
	component.Text.SetValue(b.entry, text)
	return b
}

func (b EntryBuilder) Entry() *donburi.Entry {
	return b.entry
}

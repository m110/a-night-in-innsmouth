package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewPOI(
	parent *donburi.Entry,
	pos math.Vec2,
	size engine.Size,
	passage string,
) *donburi.Entry {
	w := parent.World

	img := ebiten.NewImage(size.Width, size.Height)
	vector.StrokeRect(img, 0, 0, float32(size.Width), float32(size.Height), 10, colornames.Red, false)

	poi := NewTagged(w, "POI").
		WithParent(parent).
		WithPosition(pos).
		WithLayer(component.SpriteLayerPOI).
		WithSprite(component.SpriteData{
			Image: img,
		}).
		WithSpriteBounds().
		WithBoundsAsCollider(component.CollisionLayerPOI).
		With(component.POI).
		Entry()

	component.POI.SetValue(poi, component.POIData{
		Passage: passage,
	})

	indicatorImg := ebiten.NewImage(64, 64)
	vector.DrawFilledCircle(indicatorImg, 32, 32, 16, colornames.Indianred, true)

	NewTagged(w, "POIIndicator").
		WithParent(poi).
		WithPosition(math.Vec2{
			X: 24,
			Y: -48,
		}).
		With(component.Active).
		WithLayerInherit().
		WithSprite(component.SpriteData{
			Image: indicatorImg,
		}).
		Entry()

	return poi
}

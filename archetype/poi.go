package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
)

func NewPOI(
	parent *donburi.Entry,
	poi domain.POI,
) *donburi.Entry {
	w := parent.World

	entry := NewTagged(w, "POI").
		WithParent(parent).
		WithPosition(poi.Rect.Position()).
		WithLayer(component.SpriteLayerPOI).
		WithBounds(poi.Rect.Size()).
		WithBoundsAsCollider(component.CollisionLayerPOI).
		With(component.POI).
		Entry()

	component.POI.SetValue(entry, component.POIData{
		Passage: poi.Passage,
	})

	width := poi.Rect.Size().Width
	height := poi.Rect.Size().Height

	indicatorImg := ebiten.NewImage(width, height)
	color := colornames.Indianred
	color.A = 100
	vector.DrawFilledCircle(indicatorImg, float32(width/2.0), float32(height/2.0), float32(width/2.0), color, true)

	NewTagged(w, "POIIndicator").
		WithParent(entry).
		WithPosition(math.Vec2{
			X: 0,
			Y: 0,
		}).
		With(component.Active).
		With(component.POIIndicator).
		WithLayerInherit().
		WithSprite(component.SpriteData{
			Image: indicatorImg,
		}).
		Entry()

	return entry
}

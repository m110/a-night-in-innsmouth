package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
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
		WithPosition(poi.TriggerRect.Position()).
		WithLayer(component.SpriteLayerPOI).
		WithBounds(poi.TriggerRect.Size()).
		WithBoundsAsCollider(component.CollisionLayerPOI).
		With(component.POI).
		Entry()

	component.POI.SetValue(entry, component.POIData{
		POI: poi,
	})

	width := poi.Rect.Size().Width
	height := poi.Rect.Size().Height

	indicatorImg := ebiten.NewImage(width, height)
	color := colornames.Indianred
	color.A = 100
	vector.DrawFilledCircle(indicatorImg, float32(width/2.0), float32(height/2.0), float32(width/2.0), color, true)

	indicator := NewTagged(w, "POIIndicator").
		WithParent(entry).
		With(component.Active).
		With(component.POIIndicator).
		WithLayerInherit().
		WithSprite(component.SpriteData{
			Image: indicatorImg,
		}).
		Entry()

	transform.SetWorldPosition(indicator, math.Vec2{
		X: poi.Rect.X,
		Y: poi.Rect.Y,
	})

	return entry
}

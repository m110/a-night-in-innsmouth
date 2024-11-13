package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
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

func HidePOIs(w donburi.World) {
	activePOI, ok := donburi.NewQuery(
		filter.Contains(
			component.ActivePOI,
		),
	).First(w)
	if ok {
		activePOI.RemoveComponent(component.ActivePOI)
	}
	indicatorsQuery := donburi.NewQuery(filter.Contains(
		component.POIIndicator,
	))

	indicatorsQuery.Each(w, func(entry *donburi.Entry) {
		component.Active.Get(entry).Active = false
	})
}

func CanSeePOI(entry *donburi.Entry) bool {
	poi := component.POI.Get(entry)
	game := component.MustFindGame(entry.World)
	passage := game.Story.PassageByTitle(poi.POI.Passage)
	return passage.ConditionsMet()
}

func ShowPOI(entry *donburi.Entry) {
	entry.AddComponent(component.ActivePOI)

	poiIndicator := engine.MustFindChildWithComponent(entry, component.POIIndicator)
	component.Active.Get(poiIndicator).Active = true
}

func CheckNextPOI(w donburi.World) {
	character := engine.MustFindWithComponent(w, component.Character)
	collider := component.Collider.Get(character)

	var nextCollisionEntry *donburi.Entry
	var nextCollision *component.Collision
	for key, collision := range collider.CollidesWith {
		if key.Layer != component.CollisionLayerPOI {
			continue
		}
		if nextCollision == nil || collision.TimesSeen > nextCollision.TimesSeen {
			nextCollision = &collision
			nextCollisionEntry = character.World.Entry(key.Other)
		}
	}

	if nextCollisionEntry != nil && CanSeePOI(nextCollisionEntry) {
		ShowPOI(nextCollisionEntry)
	}
}

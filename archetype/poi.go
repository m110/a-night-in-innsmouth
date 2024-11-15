package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/definitions"
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
		WithLayer(definitions.SpriteLayerPOI).
		WithBounds(poi.TriggerRect.Size()).
		WithBoundsAsCollider(definitions.CollisionLayerPOI).
		With(component.POI).
		Entry()

	component.POI.SetValue(entry, component.POIData{
		POI: poi,
	})

	if poi.Image != nil {
		poiImage := NewTagged(w, "POIImage").
			WithParent(entry).
			WithLayerInherit().
			WithSprite(component.SpriteData{
				Image: poi.Image,
				ColorBlendOverride: &component.ColorBlendOverride{
					Value: 0,
				},
			}).
			With(component.POIImage).
			Entry()

		transform.SetWorldPosition(poiImage, math.Vec2{
			X: poi.Rect.X,
			Y: poi.Rect.Y,
		})
	}

	return entry
}

func DeactivatePOIs(w donburi.World) {
	activePOI, ok := donburi.NewQuery(
		filter.Contains(
			component.ActivePOI,
		),
	).First(w)
	if ok {
		activePOI.RemoveComponent(component.ActivePOI)
	}
}

func CanSeePOI(entry *donburi.Entry) bool {
	poi := component.POI.Get(entry)
	game := component.MustFindGame(entry.World)

	if poi.POI.Level != nil {
		return true
	}

	passage := game.Story.PassageByTitle(poi.POI.Passage)
	return passage.ConditionsMet()
}

func ActivatePOI(entry *donburi.Entry) {
	entry.AddComponent(component.ActivePOI)
}

func CheckNextPOI(w donburi.World) {
	// TODO Simple image boards could be used differently from levels
	// Remember the character pos and return to it? If simply exit
	character := engine.MustFindWithComponent(w, component.Character)
	collider := component.Collider.Get(character)

	var nextCollisionEntry *donburi.Entry
	var nextCollision *component.Collision
	for key, collision := range collider.CollidesWith {
		if key.Layer != definitions.CollisionLayerPOI {
			continue
		}
		if nextCollision == nil || collision.TimesSeen > nextCollision.TimesSeen {
			nextCollision = &collision
			nextCollisionEntry = character.World.Entry(key.Other)
		}
	}

	if nextCollisionEntry != nil && CanSeePOI(nextCollisionEntry) {
		ActivatePOI(nextCollisionEntry)
	}
}

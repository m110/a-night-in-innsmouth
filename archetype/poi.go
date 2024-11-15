package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

var PoiVisibleDistance = engine.FloatRange{Min: 200, Max: 400}

func NewPOI(
	parent *donburi.Entry,
	poi domain.POI,
) *donburi.Entry {
	w := parent.World

	entry := NewTagged(w, "POI").
		WithParent(parent).
		WithPosition(poi.TriggerRect.Position()).
		WithLayer(domain.SpriteLayerPOI).
		WithBounds(poi.TriggerRect.Size()).
		WithBoundsAsCollider(domain.CollisionLayerPOI).
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

func CanInteractWithPOI(entry *donburi.Entry) bool {
	poi := component.POI.Get(entry)

	poiImage, ok := transform.FindChildWithComponent(entry, component.POIImage)
	if ok {
		character, found := engine.FindWithComponent(entry.World, component.Character)
		if found {
			// TODO Probably should consider width + height and calculate off center
			// TODO Could be based on just X pos
			characterPos := transform.WorldPosition(character)
			poiPos := transform.WorldPosition(poiImage)
			dist := characterPos.Distance(poiPos)
			if dist > PoiVisibleDistance.Max {
				return false
			}
		}
	}

	game := component.MustFindGame(entry.World)

	if poi.POI.Level != nil {
		return true
	}

	passage := game.Story.PassageByTitle(poi.POI.Passage)
	return passage.ConditionsMet()
}

func SelectPOI(entry *donburi.Entry) {
	if !CanInteractWithPOI(entry) {
		return
	}

	poi := component.POI.Get(entry)
	game := component.MustFindGame(entry.World)

	RotateCharacterTowards(entry)

	if poi.POI.Passage != "" {
		passage := game.Story.PassageByTitle(poi.POI.Passage)
		passage.Visit()
		ShowPassage(entry.World, passage, entry)
	} else if poi.POI.Level != nil {
		ChangeLevel(entry.World, *poi.POI.Level)
	}
}

func ActivatePOI(entry *donburi.Entry) {
	entry.AddComponent(component.ActivePOI)
}

func CheckNextPOI(w donburi.World) {
	character := engine.MustFindWithComponent(w, component.Character)
	collider := component.Collider.Get(character)

	var nextCollisionEntry *donburi.Entry
	var nextCollision *component.Collision
	for key, collision := range collider.CollidesWith {
		if key.Layer != domain.CollisionLayerPOI {
			continue
		}
		if nextCollision == nil || collision.TimesSeen > nextCollision.TimesSeen {
			nextCollision = &collision
			nextCollisionEntry = character.World.Entry(key.Other)
		}
	}

	if nextCollisionEntry != nil && CanInteractWithPOI(nextCollisionEntry) {
		ActivatePOI(nextCollisionEntry)
	}
}

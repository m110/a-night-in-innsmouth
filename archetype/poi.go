package archetype

import (
	math2 "math"
	"time"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

var PoiVisibleDistance = engine.FloatRange{Min: 200, Max: 500}

func NewPOI(
	parent *donburi.Entry,
	poi domain.POI,
) *donburi.Entry {
	w := parent.World

	layer := domain.SpriteLayerObjects
	if poi.Object.Layer != 0 {
		layer = poi.Object.Layer
	}

	entry := NewTagged(w, "POI").
		WithParent(parent).
		WithPosition(poi.TriggerRect.Position()).
		WithLayer(layer).
		WithBounds(poi.TriggerRect.Size()).
		WithBoundsAsCollider(domain.CollisionLayerPOI).
		With(component.Animator).
		With(component.POI).
		Entry()

	component.POI.SetValue(entry, component.POIData{
		POI: poi,
	})

	var poiSprite *component.SpriteData
	if poi.Object.Image != nil {
		poiImage := NewTagged(w, "POIImage").
			WithParent(entry).
			WithLayerInherit().
			WithScale(poi.Object.Scale).
			WithSprite(component.SpriteData{
				Image: poi.Object.Image,
				ColorBlendOverride: &component.ColorBlendOverride{
					Value: 0,
				},
			}).
			WithSpriteBounds().
			With(component.POIImage).
			Entry()

		transform.SetWorldPosition(poiImage, math.Vec2{
			X: poi.Object.Position.X,
			Y: poi.Object.Position.Y,
		})

		poiSprite = component.Sprite.Get(poiImage)
		poiSprite.AlphaOverride = &component.AlphaOverride{
			A: 1,
		}
	}

	hidden := false

	if poi.ParentObject != nil {
		obj := NewObject(parent, *poi.ParentObject)
		component.POI.Get(entry).ParentObject = obj

		parentObjectSprite := component.Sprite.Get(obj)
		parentObjectSprite.AlphaOverride = &component.AlphaOverride{
			A: 1,
		}

		anim := component.Animator.Get(entry)
		anim.SetAnimation("fade-in", &component.Animation{
			Timer: engine.NewTimer(time.Second),
			OnStart: func(e *donburi.Entry) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = 0
				}
				parentObjectSprite.AlphaOverride.A = 0
				hidden = false
			},
			OnStop: func(e *donburi.Entry) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = 1
				}
				parentObjectSprite.AlphaOverride.A = 1
			},
			Update: func(e *donburi.Entry, a *component.Animation) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = a.Timer.PercentDone()
				}
				parentObjectSprite.AlphaOverride.A = a.Timer.PercentDone()
				a.Timer.Update()
				if a.Timer.IsReady() {
					a.Stop(entry)
				}
			},
		})

		anim.SetAnimation("fade-out", &component.Animation{
			Timer: engine.NewTimer(time.Second),
			OnStart: func(e *donburi.Entry) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = 1
				}
				parentObjectSprite.AlphaOverride.A = 1
				hidden = true
			},
			OnStop: func(e *donburi.Entry) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = 0
				}
				parentObjectSprite.AlphaOverride.A = 0
			},
			Update: func(e *donburi.Entry, a *component.Animation) {
				if poiSprite != nil {
					poiSprite.AlphaOverride.A = 1 - a.Timer.PercentDone()
				}
				parentObjectSprite.AlphaOverride.A = 1 - a.Timer.PercentDone()
				a.Timer.Update()
				if a.Timer.IsReady() {
					a.Stop(entry)
				}
			},
		})

		if !POIConditionsMet(parent.World, poi) {
			hidden = true
			if poiSprite != nil {
				poiSprite.AlphaOverride.A = 0
			}
			parentObjectSprite.AlphaOverride.A = 0
		}

		checkConditions := func() {
			conditionsMet := POIConditionsMet(parent.World, poi)
			if hidden && conditionsMet {
				anim.Start("fade-in", entry)
			} else if !hidden && !conditionsMet {
				anim.Start("fade-out", entry)
			}
		}

		domain.StoryFactSetEvent.Subscribe(w, func(w donburi.World, event domain.StoryFactSet) {
			checkConditions()
		})

		domain.InventoryUpdatedEvent.Subscribe(w, func(w donburi.World, event domain.InventoryUpdated) {
			checkConditions()
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
	poiImage, ok := transform.FindChildWithComponent(entry, component.POIImage)
	if ok {
		character, found := engine.FindWithComponent(entry.World, component.Character)
		if found {
			characterPos := HorizontalCenterPosition(character)
			poiPos := HorizontalCenterPosition(poiImage)
			dist := math2.Abs(characterPos - poiPos)
			if dist > PoiVisibleDistance.Max {
				return false
			}
		}
	}

	poi := component.POI.Get(entry)
	return POIConditionsMet(entry.World, poi.POI)
}

func POIConditionsMet(w donburi.World, poi domain.POI) bool {
	game := component.MustFindGame(w)

	if poi.Level != nil {
		passage, ok := game.Story.PassageForLevel(*poi.Level)
		if ok {
			return passage.ConditionsMet()
		}
		return true
	}

	passage := game.Story.PassageByTitle(poi.Passage)
	return passage.ConditionsMet()
}

func HorizontalCenterPosition(entry *donburi.Entry) float64 {
	pos := transform.WorldPosition(entry)
	bounds := component.Bounds.Get(entry)
	return pos.X + bounds.Width/2
}

func SelectPOI(entry *donburi.Entry) bool {
	if !CanInteractWithPOI(entry) {
		return false
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

	return true
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

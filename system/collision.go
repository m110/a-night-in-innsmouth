package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
	"github.com/m110/secrets/events"
)

type Collision struct {
	query *donburi.Query
}

func NewCollision() *Collision {
	return &Collision{
		query: donburi.NewQuery(filter.Contains(component.Collider)),
	}
}

func (c *Collision) Init(w donburi.World) {
	events.JustCollidedEvent.Subscribe(w, func(w donburi.World, event events.JustCollided) {
		// TODO remove casting
		if event.Layer == int(component.CollisionLayerCharacter) && event.OtherLayer == int(component.CollisionLayerPOI) {
			hidePOIs(w)

			event.Other.AddComponent(component.ActivePOI)

			poiIndicator := engine.MustFindChildWithComponent(event.Other, component.POIIndicator)
			component.Active.Get(poiIndicator).Active = true
		}
	})

	events.JustOutOfCollisionEvent.Subscribe(w, func(w donburi.World, event events.JustOutOfCollision) {
		if event.Layer == int(component.CollisionLayerCharacter) && event.OtherLayer == int(component.CollisionLayerPOI) {
			if event.Other.HasComponent(component.ActivePOI) {
				hidePOIs(w)
				// TODO Check active collisions and switch POI if found
			}
		}
	})
}

func hidePOIs(w donburi.World) {
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

var collisions = map[component.ColliderLayer]map[component.ColliderLayer]struct{}{
	component.CollisionLayerCharacter: {
		component.CollisionLayerPOI: {},
	},
}

func (c *Collision) Update(w donburi.World) {
	var entries []*donburi.Entry
	c.query.Each(w, func(entry *donburi.Entry) {
		if !entry.Valid() {
			return
		}

		entries = append(entries, entry)
	})

	for _, entry := range entries {
		collider := component.Collider.Get(entry)
		for i, c := range collider.CollidesWith {
			c.Detected = false
			collider.CollidesWith[i] = c
		}
		collider.JustCollidedWith = nil
		collider.JustOutOfCollisionWith = nil
	}

	for _, entry := range entries {
		collider := component.Collider.Get(entry)

		for _, other := range entries {
			if entry.Entity().Id() == other.Entity().Id() {
				continue
			}

			otherCollider := component.Collider.Get(other)

			_, ok := collisions[collider.Layer][otherCollider.Layer]
			if !ok {
				continue
			}

			pos := transform.Transform.Get(entry).LocalPosition
			otherPos := transform.Transform.Get(other).LocalPosition

			// TODO The current approach doesn't take rotation into account
			// TODO The current approach doesn't take scale into account
			rect := engine.NewRect(pos.X, pos.Y, collider.Width, collider.Height)
			otherRect := engine.NewRect(otherPos.X, otherPos.Y, otherCollider.Width, otherCollider.Height)

			if rect.Intersects(otherRect) {
				key := component.CollisionKey{
					Layer: otherCollider.Layer,
					Other: other.Entity(),
				}

				currentCollision, ok := collider.CollidesWith[key]
				if !ok {
					if collider.JustCollidedWith == nil {
						collider.JustCollidedWith = map[component.CollisionKey]struct{}{}
					}

					collider.JustCollidedWith[key] = struct{}{}

					event := events.JustCollided{
						Entry:      entry,
						Layer:      int(collider.Layer),
						Other:      other,
						OtherLayer: int(otherCollider.Layer),
					}
					events.JustCollidedEvent.Publish(w, event)
				}

				currentCollision.Detected = true
				currentCollision.TimesSeen++

				if collider.CollidesWith == nil {
					collider.CollidesWith = map[component.CollisionKey]component.Collision{}
				}

				collider.CollidesWith[key] = currentCollision
			}
		}
	}

	for _, entry := range entries {
		collider := component.Collider.Get(entry)

		for key, collision := range collider.CollidesWith {
			if !collision.Detected {
				if collider.JustOutOfCollisionWith == nil {
					collider.JustOutOfCollisionWith = map[component.CollisionKey]struct{}{}
				}

				collider.JustOutOfCollisionWith[key] = struct{}{}
				delete(collider.CollidesWith, key)

				event := events.JustOutOfCollision{
					Entry:      entry,
					Layer:      int(collider.Layer),
					Other:      w.Entry(key.Other),
					OtherLayer: int(key.Layer),
				}
				events.JustOutOfCollisionEvent.Publish(w, event)
			}
		}
	}
}

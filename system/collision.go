package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/archetype"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

type Collision struct {
	query *donburi.Query
}

func NewCollision() *Collision {
	return &Collision{
		query: donburi.NewQuery(filter.Contains(component.Collider)),
	}
}

type collisionEffect func(w donburi.World, entry *donburi.Entry, other *donburi.Entry)

var collisionEffects = map[component.ColliderLayer]map[component.ColliderLayer]collisionEffect{
	component.CollisionLayerCharacter: {
		component.CollisionLayerPOI: func(w donburi.World, entry *donburi.Entry, other *donburi.Entry) {
			indicatorsQuery := donburi.NewQuery(filter.Contains(
				component.POIIndicator,
			))

			indicatorsQuery.Each(w, func(entry *donburi.Entry) {
				component.Active.Get(entry).Active = false
			})

			dialog := engine.MustFindWithComponent(w, component.Dialog)
			if component.Active.Get(dialog).Active {
				return
			}

			game := component.MustFindGame(w)

			poi := component.POI.Get(other)
			archetype.ShowPassage(w, game.Story.PassageByTitle(poi.Passage))

			//poiIndicator := engine.MustFindChildWithComponent(other, component.POIIndicator)
			//component.Active.Get(poiIndicator).Active = true
		},
	},
}

func (c *Collision) Update(w donburi.World) {
	var entries []*donburi.Entry
	c.query.Each(w, func(entry *donburi.Entry) {
		entries = append(entries, entry)
	})

	for _, entry := range entries {
		if !entry.Valid() {
			continue
		}

		collider := component.Collider.Get(entry)

		for _, other := range entries {
			if entry.Entity().Id() == other.Entity().Id() {
				continue
			}

			otherCollider := component.Collider.Get(other)

			effects, ok := collisionEffects[collider.Layer]
			if !ok {
				continue
			}

			effect, ok := effects[otherCollider.Layer]
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
				effect(w, entry, other)
			}
		}
	}
}

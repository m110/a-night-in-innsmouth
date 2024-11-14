package events

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"

	"github.com/m110/secrets/definitions"
)

type Item struct {
	Name  string
	Count int
}

type InventoryUpdated struct {
	Money int
	Items []Item
}

var InventoryUpdatedEvent = events.NewEventType[InventoryUpdated]()

type JustCollided struct {
	Entry      *donburi.Entry
	Layer      definitions.ColliderLayer
	Other      *donburi.Entry
	OtherLayer definitions.ColliderLayer
}

var JustCollidedEvent = events.NewEventType[JustCollided]()

type JustOutOfCollision struct {
	Entry      *donburi.Entry
	Layer      definitions.ColliderLayer
	Other      *donburi.Entry
	OtherLayer definitions.ColliderLayer
}

var JustOutOfCollisionEvent = events.NewEventType[JustOutOfCollision]()

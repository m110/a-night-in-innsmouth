package events

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
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
	Layer      int
	Other      *donburi.Entry
	OtherLayer int
}

var JustCollidedEvent = events.NewEventType[JustCollided]()

type JustOutOfCollision struct {
	Entry      *donburi.Entry
	Layer      int
	Other      *donburi.Entry
	OtherLayer int
}

var JustOutOfCollisionEvent = events.NewEventType[JustOutOfCollision]()

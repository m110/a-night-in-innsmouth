package events

import "github.com/yohamta/donburi/features/events"

type Item struct {
	Name  string
	Count int
}

type InventoryUpdated struct {
	Money int
	Items []Item
}

var InventoryUpdatedEvent = events.NewEventType[InventoryUpdated]()

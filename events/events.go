package events

import "github.com/yohamta/donburi/features/events"

type Item struct {
	Name  string
	Count int
}

type InventoryUpdated struct {
	Items []Item
}

var InventoryUpdatedEvent = events.NewEventType[InventoryUpdated]()

type MoneyUpdated struct {
	Amount int
}

var MoneyUpdatedEvent = events.NewEventType[MoneyUpdated]()

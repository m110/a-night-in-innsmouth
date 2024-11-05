package events

import "github.com/yohamta/donburi/features/events"

type MoneyUpdated struct {
	Amount int
}

var MoneyUpdatedEvent = events.NewEventType[MoneyUpdated]()

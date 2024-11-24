package domain

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
)

type InventoryItem struct {
	Name  string
	Count int
}

type InventoryUpdated struct {
	Money int
	Items []InventoryItem
}

var InventoryUpdatedEvent = events.NewEventType[InventoryUpdated]()

type ItemReceived struct {
	Item InventoryItem
}

var ItemReceivedEvent = events.NewEventType[ItemReceived]()

type ItemLost struct {
	Item InventoryItem
}

var ItemLostEvent = events.NewEventType[ItemLost]()

type MoneyReceived struct {
	Amount int
}

var MoneyReceivedEvent = events.NewEventType[MoneyReceived]()

type MoneySpent struct {
	Amount int
}

var MoneySpentEvent = events.NewEventType[MoneySpent]()

type JustCollided struct {
	Entry      *donburi.Entry
	Layer      ColliderLayer
	Other      *donburi.Entry
	OtherLayer ColliderLayer
}

var JustCollidedEvent = events.NewEventType[JustCollided]()

type JustOutOfCollision struct {
	Entry      *donburi.Entry
	Layer      ColliderLayer
	Other      *donburi.Entry
	OtherLayer ColliderLayer
}

var JustOutOfCollisionEvent = events.NewEventType[JustOutOfCollision]()

type ButtonClicked struct{}

var ButtonClickedEvent = events.NewEventType[ButtonClicked]()

type MusicChanged struct {
	Track string
}

var MusicChangedEvent = events.NewEventType[MusicChanged]()

type CharacterSpeedChanged struct {
	SpeedChange float64
}

var CharacterSpeedChangedEvent = events.NewEventType[CharacterSpeedChanged]()

type StoryFactSet struct {
	Fact string
}

var StoryFactSetEvent = events.NewEventType[StoryFactSet]()

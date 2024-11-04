package component

import "github.com/yohamta/donburi"

var (
	// UI marks the UI root entity.
	// Other UI elements should be children of this entity.
	UI = donburi.NewTag()
)

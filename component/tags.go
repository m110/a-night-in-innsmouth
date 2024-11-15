package component

import (
	"github.com/yohamta/donburi"
)

var (
	ActiveOptionIndicator = donburi.NewTag()

	Dialog          = donburi.NewTag()
	DialogCamera    = donburi.NewTag()
	DialogLog       = donburi.NewTag()
	DialogLogCamera = donburi.NewTag()

	// InnerCamera means the camera is placed inside another camera.
	InnerCamera = donburi.NewTag()

	Level       = donburi.NewTag()
	LevelCamera = donburi.NewTag()
	Character   = donburi.NewTag()

	Inventory = donburi.NewTag()

	ActivePOI = donburi.NewTag()
	POIImage  = donburi.NewTag()
)

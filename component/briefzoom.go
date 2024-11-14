package component

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
)

type BriefZoomData struct {
	OriginalPosition math.Vec2
	OriginalZoom     float64
	Source           *donburi.Entry
}

var BriefZoom = donburi.NewComponentType[BriefZoomData]()

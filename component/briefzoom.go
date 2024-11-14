package component

import (
	"github.com/yohamta/donburi"
)

type BriefZoomData struct {
	OriginCamera CameraData
}

var BriefZoom = donburi.NewComponentType[BriefZoomData]()

package component

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type CameraData struct {
	Viewport *ebiten.Image
	Root     *donburi.Entry
	Index    int
}

func (d CameraData) Order() int {
	return d.Index
}

var Camera = donburi.NewComponentType[CameraData]()

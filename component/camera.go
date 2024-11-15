package component

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/engine"
)

type CameraData struct {
	Viewport         *ebiten.Image
	ViewportPosition math.Vec2
	ViewportZoom     float64
	ViewportTarget   *donburi.Entry
	ViewportBounds   ViewportBounds

	Root  *donburi.Entry
	Index int
	Mask  *ebiten.Image

	TransitionOverlay *ebiten.Image
	TransitionAlpha   float64

	ColorOverride *ColorOverride
	AlphaOverride *AlphaOverride
}

type ViewportBounds struct {
	X *engine.FloatRange
	Y *engine.FloatRange
}

func (d *CameraData) SetViewportPosition(pos math.Vec2) {
	if d.ViewportBounds.X != nil {
		pos.X = engine.Clamp(pos.X, d.ViewportBounds.X.Min, d.ViewportBounds.X.Max)
	}

	if d.ViewportBounds.Y != nil {
		pos.Y = engine.Clamp(pos.Y, d.ViewportBounds.Y.Min, d.ViewportBounds.Y.Max)
	}

	d.ViewportPosition = pos
}

func (d CameraData) Order() int {
	return d.Index
}

func (d *CameraData) WorldPositionToViewportPosition(e *donburi.Entry) math.Vec2 {
	pos := transform.WorldPosition(e)
	pos = pos.Sub(d.ViewportPosition)
	pos = pos.MulScalar(d.ViewportZoom)
	return pos
}

var Camera = donburi.NewComponentType[CameraData]()

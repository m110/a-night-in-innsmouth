package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewCamera(
	w donburi.World,
	startPosition math.Vec2,
	dimensions engine.Size,
	index int,
	root *donburi.Entry,
) *donburi.Entry {
	camera := New(w).
		WithPosition(startPosition).
		With(component.Camera).
		Entry()

	viewport := ebiten.NewImage(dimensions.Width, dimensions.Height)

	component.Camera.SetValue(camera, component.CameraData{
		Viewport: viewport,
		Root:     root,
		Index:    index,
	})

	return camera
}

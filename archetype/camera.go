package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewCamera(w donburi.World, startPosition math.Vec2, zoom engine.FloatRange) *donburi.Entry {
	camera := New(w).
		WithPosition(startPosition).
		With(component.Camera).
		Entry()

	component.Camera.SetValue(camera, component.CameraData{
		Zoom: zoom,
	})

	return camera
}

func MustFindCamera(w donburi.World) *donburi.Entry {
	return engine.MustFindWithComponent(w, component.Camera)
}

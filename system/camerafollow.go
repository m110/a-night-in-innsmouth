package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type CameraFollow struct {
	query *donburi.Query
}

func NewCameraFollow() *CameraFollow {
	return &CameraFollow{
		query: donburi.NewQuery(
			filter.Contains(
				component.Camera,
			),
		),
	}
}

func (s *CameraFollow) Update(w donburi.World) {
	s.query.Each(w, func(entry *donburi.Entry) {
		cam := component.Camera.Get(entry)
		if cam.ViewportTarget == nil {
			return
		}

		pos := transform.WorldPosition(cam.ViewportTarget)

		cam.ViewportPosition.X = pos.X - float64(cam.Viewport.Bounds().Dx())/2

		if cam.ViewportBounds != nil {
			if cam.ViewportPosition.X < cam.ViewportBounds.X {
				cam.ViewportPosition.X = cam.ViewportBounds.X
			} else if cam.ViewportPosition.X+float64(cam.Viewport.Bounds().Dx()) > cam.ViewportBounds.X+cam.ViewportBounds.Width {
				cam.ViewportPosition.X = cam.ViewportBounds.X + cam.ViewportBounds.Width - float64(cam.Viewport.Bounds().Dx())
			}
		}
	})
}

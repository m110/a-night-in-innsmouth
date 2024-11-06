package system

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type Controls struct {
	query       *donburi.Query
	dialogQuery *donburi.Query
}

func NewControls() *Controls {
	return &Controls{
		query: donburi.NewQuery(
			filter.Contains(
				component.Input,
				component.Velocity,
			),
		),
		dialogQuery: donburi.NewQuery(
			filter.Contains(
				component.Dialog,
			),
		),
	}
}

func (i *Controls) Update(w donburi.World) {
	i.query.Each(w, func(entry *donburi.Entry) {
		in := component.Input.Get(entry)

		if in.Disabled {
			return
		}

		if i.dialogQuery.Count(w) > 0 {
			return
		}

		velocity := component.Velocity.Get(entry)

		if ebiten.IsKeyPressed(in.MoveUpKey) {
			velocity.Velocity.Y = -in.MoveSpeed
		} else if ebiten.IsKeyPressed(in.MoveDownKey) {
			velocity.Velocity.Y = in.MoveSpeed
		}

		if ebiten.IsKeyPressed(in.MoveRightKey) {
			velocity.Velocity.X = in.MoveSpeed
		}
		if ebiten.IsKeyPressed(in.MoveLeftKey) {
			velocity.Velocity.X = -in.MoveSpeed
		}
	})
}

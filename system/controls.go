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
				component.Sprite,
				component.Velocity,
				component.Animation,
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

		var stop bool
		if in.Disabled {
			stop = true
		}

		dialog, ok := i.dialogQuery.First(w)
		if ok && component.Active.Get(dialog).Active {
			stop = true
		}

		velocity := component.Velocity.Get(entry)
		sprite := component.Sprite.Get(entry)
		anim := component.Animation.Get(entry)

		if !stop {
			if ebiten.IsKeyPressed(in.MoveRightKey) {
				velocity.Velocity.X = in.MoveSpeed
				sprite.FlipY = false
				anim.Start(entry)
			} else if ebiten.IsKeyPressed(in.MoveLeftKey) {
				velocity.Velocity.X = -in.MoveSpeed
				sprite.FlipY = true
				anim.Start(entry)
			} else {
				stop = true
			}
		}

		if stop {
			velocity.Velocity.X = 0
			anim.Stop(entry)
		}
	})
}

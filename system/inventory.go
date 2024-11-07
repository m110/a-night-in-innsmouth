package system

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

type Inventory struct {
	query *donburi.Query
}

func NewInventory() *Inventory {
	return &Inventory{
		query: donburi.NewQuery(
			filter.Contains(
				component.Inventory,
			),
		),
	}
}

func (i *Inventory) Update(w donburi.World) {
	var clicked bool
	x, y := 0, 0

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		x, y = ebiten.TouchPosition(touchIDs[0])
		clicked = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y = ebiten.CursorPosition()
		clicked = true
	}

	var inventoryClicked bool
	if clicked {
		clickRect := engine.NewRect(float64(x), float64(y), 1, 1)

		i.query.Each(w, func(entry *donburi.Entry) {
			if !component.Active.Get(entry).Active {
				return
			}

			collider := component.Collider.Get(entry)
			pos := transform.WorldPosition(entry)
			colliderRect := engine.NewRect(pos.X, pos.Y, collider.Width, collider.Height)

			if colliderRect.Intersects(clickRect) {
				inventoryClicked = true
			}
		})
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyE) || inventoryClicked {
		i.query.Each(w, func(entry *donburi.Entry) {
			component.Active.Get(entry).Toggle()
		})
	}
}

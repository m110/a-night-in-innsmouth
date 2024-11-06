package system

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
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
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		i.query.Each(w, func(entry *donburi.Entry) {
			component.Active.Get(entry).Toggle()
		})
	}
}

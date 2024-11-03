package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type Destroy struct {
	query *donburi.Query
}

func NewDestroy() *Destroy {
	return &Destroy{
		query: donburi.NewQuery(
			filter.Contains(component.Destroyed),
		),
	}
}

func (d *Destroy) Update(w donburi.World) {
	d.query.Each(w, func(entry *donburi.Entry) {
		transform.RemoveRecursive(entry)
	})
}

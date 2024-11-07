package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type Animation struct {
	query *donburi.Query
}

func NewAnimation() *Animation {
	return &Animation{
		query: donburi.NewQuery(filter.Contains(component.Animation)),
	}
}

func (s *Animation) Init(w donburi.World) {}

func (s *Animation) Update(w donburi.World) {
	s.query.Each(w, func(entry *donburi.Entry) {
		animation := component.Animation.Get(entry)
		if !animation.Active {
			return
		}
		if animation.Timer != nil {
			animation.Timer.Update()
		}
		animation.Update(entry)
	})
}

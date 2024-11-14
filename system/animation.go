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
		query: donburi.NewQuery(filter.Contains(component.Animator)),
	}
}

func (s *Animation) Init(w donburi.World) {}

func (s *Animation) Update(w donburi.World) {
	s.query.Each(w, func(entry *donburi.Entry) {
		animator := component.Animator.Get(entry)
		for _, animation := range animator.Animations {
			if !animation.Active {
				continue
			}
			if animation.Timer != nil {
				animation.Timer.Update()
			}
			animation.Update(entry, animation)
		}
	})
}

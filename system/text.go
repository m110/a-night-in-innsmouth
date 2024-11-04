package system

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type Text struct {
	query *donburi.Query
}

func NewText() *Text {
	return &Text{
		query: donburi.NewQuery(
			filter.Contains(
				transform.Transform,
				component.Text,
			),
		),
	}
}

func (t *Text) Init(w donburi.World) {}

func (t *Text) Update(w donburi.World) {
	t.query.Each(w, func(entry *donburi.Entry) {
		txt := component.Text.Get(entry)
		if txt.Streaming && !txt.StreamingTimer.IsReady() {
			txt.StreamingTimer.Update()
		}
	})
}

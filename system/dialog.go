package system

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

type Dialog struct {
	query *donburi.Query
}

func NewDialog() *Dialog {
	return &Dialog{
		query: donburi.NewQuery(filter.Contains(component.Dialog)),
	}
}

func (d *Dialog) Update(w donburi.World) {
	d.query.Each(w, func(entry *donburi.Entry) {
		entry, ok := d.query.First(w)
		if !ok {
			return
		}

		dialog := component.Dialog.Get(entry)

		// Game over?
		if len(dialog.Passage.Links()) == 0 {
			return
		}

		var updated bool
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			dialog.ActiveOption++
			updated = true
		} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			dialog.ActiveOption--
			updated = true
		}

		if dialog.ActiveOption < 0 {
			dialog.ActiveOption = len(dialog.Passage.Links()) - 1
		}

		if dialog.ActiveOption >= len(dialog.Passage.Links()) {
			dialog.ActiveOption = 0
		}

		if updated {
			indicator := engine.MustFindWithComponent(w, component.ActiveOptionIndicator)
			dialogOptions := engine.FindChildrenWithComponent(entry, component.DialogOption)

			transform.ChangeParent(indicator, dialogOptions[dialog.ActiveOption], false)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			link := dialog.Passage.Links()[dialog.ActiveOption]

			link.Visit()

			transform.RemoveRecursive(entry)
			archetype.NewDialog(w, link.Target)
		}
	})
}

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
	query        *donburi.Query
	buttonsQuery *donburi.Query
}

func NewDialog() *Dialog {
	return &Dialog{
		query: donburi.NewQuery(filter.Contains(component.Dialog)),
		buttonsQuery: donburi.NewQuery(
			filter.Contains(
				component.Collider,
				component.DialogOption,
			),
		),
	}
}

func (d *Dialog) Update(w donburi.World) {
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

	x, y := ebiten.CursorPosition()

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	touched := len(touchIDs) > 0
	if touched {
		x, y = ebiten.TouchPosition(touchIDs[0])
	}

	clickRect := engine.NewRect(float64(x), float64(y), 1, 1)

	d.buttonsQuery.Each(w, func(entry *donburi.Entry) {
		pos := transform.WorldPosition(entry)
		collider := component.Collider.Get(entry)
		colliderRect := engine.NewRect(pos.X, pos.Y, collider.Width, collider.Height)
		if colliderRect.Intersects(clickRect) {
			dialog.ActiveOption = component.DialogOption.Get(entry).Index
			updated = true
		}
	})

	if updated {
		if dialog.ActiveOption < 0 {
			dialog.ActiveOption = len(dialog.Passage.Links()) - 1
		}

		if dialog.ActiveOption >= len(dialog.Passage.Links()) {
			dialog.ActiveOption = 0
		}

		indicator := engine.MustFindWithComponent(w, component.ActiveOptionIndicator)
		dialogOptions := engine.FindChildrenWithComponent(entry, component.DialogOption)

		transform.ChangeParent(indicator, dialogOptions[dialog.ActiveOption], false)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && updated) ||
		(touched && updated) {
		link := dialog.Passage.Links()[dialog.ActiveOption]

		link.Visit()

		transform.RemoveRecursive(entry)
		archetype.NewDialog(w, link.Target)
	}
}

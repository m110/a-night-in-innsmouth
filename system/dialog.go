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
		query: donburi.NewQuery(filter.Contains(component.Passage)),
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

	if !isActive(entry) {
		return
	}

	passage := component.Passage.Get(entry)
	stack := engine.MustGetParent(entry)
	stackedView := component.StackedView.Get(stack)

	// Game over?
	if len(passage.Passage.Links()) == 0 {
		return
	}

	var updated bool
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		passage.ActiveOption++
		updated = true
	} else if inpututil.IsKeyJustPressed(ebiten.KeyUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		passage.ActiveOption--
		updated = true
	}

	_, wy := ebiten.Wheel()
	var scroll int
	if ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		scroll = 10
	} else if ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		scroll = -10
	} else if wy < 0 {
		scroll = -25
	} else if wy > 0 {
		scroll = 25
	}

	var touched bool
	if scroll == 0 {
		x, y := ebiten.CursorPosition()

		touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
		touched = len(touchIDs) > 0
		if touched {
			x, y = ebiten.TouchPosition(touchIDs[0])
		}
		clickRect := engine.NewRect(float64(x), float64(y), 1, 1)

		d.buttonsQuery.Each(w, func(entry *donburi.Entry) {
			pos := transform.WorldPosition(entry)
			collider := component.Collider.Get(entry)
			colliderRect := engine.NewRect(pos.X, pos.Y, collider.Width, collider.Height)
			if colliderRect.Intersects(clickRect) {
				passage.ActiveOption = component.DialogOption.Get(entry).Index
				updated = true
			}
		})
	}

	if updated {
		if passage.ActiveOption < 0 {
			passage.ActiveOption = len(passage.Passage.Links()) - 1
		}

		if passage.ActiveOption >= len(passage.Passage.Links()) {
			passage.ActiveOption = 0
		}

		indicator := engine.MustFindWithComponent(w, component.ActiveOptionIndicator)
		dialogOptions := engine.FindChildrenWithComponent(entry, component.DialogOption)

		transform.ChangeParent(indicator, dialogOptions[passage.ActiveOption], false)
	}

	var next bool
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		(inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && updated) ||
		(touched && updated) {
		next = true
	}

	if next && !stackedView.Scrolled {
		archetype.NextPassage(w)
	}

	stackTransform := transform.GetTransform(stack)

	if updated || next {
		stackTransform.LocalPosition.Y = -stackedView.CurrentY
		stackedView.Scrolled = false
	} else if scroll != 0 {
		stackedView.Scrolled = true
		stackTransform.LocalPosition.Y += float64(scroll)

		if stackTransform.LocalPosition.Y > 0 {
			stackTransform.LocalPosition.Y = 0
		}

		if stackTransform.LocalPosition.Y <= -stackedView.CurrentY {
			stackTransform.LocalPosition.Y = -stackedView.CurrentY
			stackedView.Scrolled = false
		}
	}
}

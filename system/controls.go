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

type Controls struct {
	characterQuery *donburi.Query
	dialogQuery    *donburi.Query
	activePOIQuery *donburi.Query
	passageQuery   *donburi.Query
	buttonsQuery   *donburi.Query
}

func NewControls() *Controls {
	return &Controls{
		characterQuery: donburi.NewQuery(
			filter.Contains(
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
		activePOIQuery: donburi.NewQuery(
			filter.Contains(
				component.ActivePOI,
			),
		),
		passageQuery: donburi.NewQuery(filter.Contains(component.Passage)),
		buttonsQuery: donburi.NewQuery(
			filter.Contains(
				component.Collider,
				component.DialogOption,
			),
		),
	}
}

func (c *Controls) Update(w donburi.World) {
	in := engine.MustFindComponent[component.InputData](w, component.Input)

	if in.Disabled {
		return
	}

	character, characterFound := c.characterQuery.First(w)

	dialog, ok := c.dialogQuery.First(w)
	if ok && component.Active.Get(dialog).Active {
		if characterFound {
			stopCharacter(character)
		}
		c.UpdateDialog(w)
		return
	}

	if !characterFound {
		return
	}

	velocity := component.Velocity.Get(character)
	anim := component.Animation.Get(character)

	sprite := component.Sprite.Get(character)

	var moving bool
	if ebiten.IsKeyPressed(in.MoveRightKey) {
		velocity.Velocity.X = in.MoveSpeed
		sprite.FlipY = false
		anim.Start(character)
		moving = true
	} else if ebiten.IsKeyPressed(in.MoveLeftKey) {
		velocity.Velocity.X = -in.MoveSpeed
		sprite.FlipY = true
		anim.Start(character)
		moving = true
	}

	if !moving {
		velocity.Velocity.X = 0
		anim.Stop(character)
	}

	if inpututil.IsKeyJustPressed(in.ActionKey) {
		activePOI, ok := c.activePOIQuery.First(w)
		if ok {
			game := component.MustFindGame(w)
			poi := component.POI.Get(activePOI)

			if poi.POI.Passage != "" {
				archetype.ShowPassage(w, game.Story.PassageByTitle(poi.POI.Passage))
			} else if poi.POI.Level != nil {
				archetype.ChangeLevel(w, *poi.POI.Level)
			}
		}
	}
}

func stopCharacter(entry *donburi.Entry) {
	velocity := component.Velocity.Get(entry)
	anim := component.Animation.Get(entry)
	velocity.Velocity.X = 0
	anim.Stop(entry)
}

func (c *Controls) UpdateDialog(w donburi.World) {
	entry, ok := c.passageQuery.First(w)
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

		c.buttonsQuery.Each(w, func(entry *donburi.Entry) {
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
		c.buttonsQuery.Each(w, func(entry *donburi.Entry) {
			if component.DialogOption.Get(entry).Index == passage.ActiveOption {
				transform.ChangeParent(indicator, entry, false)
			}
		})
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

	camera := component.Camera.Get(engine.MustFindWithComponent(w, component.DialogCamera))

	if updated || next {
		camera.ViewportPosition.Y = stackedView.CurrentY
		stackedView.Scrolled = false
	} else if scroll != 0 {
		stackedView.Scrolled = true
		camera.ViewportPosition.Y -= float64(scroll)

		// TODO Could use a "boundary" on Camera to prevent going out of bounds
		if camera.ViewportPosition.Y < 0 {
			camera.ViewportPosition.Y = 0
		}

		if camera.ViewportPosition.Y >= stackedView.CurrentY {
			camera.ViewportPosition.Y = stackedView.CurrentY
			stackedView.Scrolled = false
		}
	}
}

package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

const (
	levelMovementMargin               = 100
	levelCameraMarginPercent          = 0.2
	scrollingLevelCameraMarginPercent = 0.45
)

func NewLevel(w donburi.World, targetLevel domain.TargetLevel) {
	level, ok := assets.Assets.Levels[targetLevel.Name]
	if !ok {
		panic("Name not found: " + targetLevel.Name)
	}

	background := level.Background()

	entry := NewTagged(w, "Level").
		WithLayer(domain.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: background,
		}).
		With(component.Level).
		With(component.Animator).
		Entry()

	component.Level.SetValue(entry, component.LevelData{
		Name: targetLevel.Name,
	})

	spawned := false

	levelCam := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCam)
	input := engine.MustFindComponent[component.InputData](w, component.Input)

	anim := component.Animator.Get(entry)

	anim.SetAnimation("transition", &component.Animation{
		Active: true,
		Timer:  engine.NewTimer(LevelTransitionDuration),
		OnStart: func(e *donburi.Entry) {
			input.Disabled = true
		},
		Update: func(e *donburi.Entry, a *component.Animation) {
			if spawned {
				cam.TransitionAlpha = a.Timer.PercentDone()
				if a.Timer.IsReady() {
					a.Stop(entry)
				}
			} else {
				cam.TransitionAlpha = 1 - a.Timer.PercentDone()
				if a.Timer.IsReady() {
					spawned = true
					input.Disabled = false
					a.Stop(entry)
				}
			}
		},
	})

	for _, poi := range level.POIs {
		NewPOI(entry, poi)
	}

	for _, o := range level.Objects {
		NewObject(entry, o)
	}

	game := component.MustFindGame(w)

	if level.StartPassage != "" {
		ShowPassage(w, game.Story.PassageByTitle(level.StartPassage), nil)
	}

	var character *donburi.Entry

	var characterPos *domain.CharacterPosition
	if len(level.Entrypoints) > 0 && targetLevel.Entrypoint != nil {
		entrypoint := level.Entrypoints[*targetLevel.Entrypoint]
		characterPos = &entrypoint.CharacterPosition

		// For now, all levels have only one Y position for the character
		// For convenience, the first entrypoint's Y position is used
		firstEntrypoint := level.Entrypoints[0]
		characterPos.LocalPosition.Y = firstEntrypoint.CharacterPosition.LocalPosition.Y
	} else if targetLevel.CharacterPosition != nil {
		characterPos = targetLevel.CharacterPosition
	}

	if characterPos != nil {
		// Default to the background boundaries
		boundsRange := engine.FloatRange{
			Min: levelMovementMargin,
			Max: float64(background.Bounds().Dx() - levelMovementMargin),
		}

		if level.Limits != nil {
			boundsRange = *level.Limits
		}

		bounds := component.MovementBoundsData{
			Range: boundsRange,
		}

		character = NewCharacter(entry, level.CharacterScale, bounds)

		transform.GetTransform(character).LocalPosition = characterPos.LocalPosition
		component.Sprite.Get(character).FlipY = characterPos.FlipY
	}

	cam.Root = entry

	if level.CameraZoom != 0 {
		cam.ViewportZoom = level.CameraZoom
	} else {
		// Calculate zoom to fit height with margins
		marginPercent := 0.01
		screenHeight := float64(game.Dimensions.ScreenHeight)
		bgHeight := float64(background.Bounds().Dy())

		totalMarginHeight := screenHeight * marginPercent * 2
		availableHeight := screenHeight - totalMarginHeight

		cam.ViewportZoom = availableHeight / bgHeight
	}

	// Multiply by zoom to go from world space to screen space
	// Divide by zoom to go from screen space to world space
	heightDiff := (float64(game.Dimensions.ScreenHeight) - float64(background.Bounds().Dy())*cam.ViewportZoom) / cam.ViewportZoom
	if heightDiff > 0 {
		cam.ViewportPosition.Y = -heightDiff / 2
	} else {
		// Should not happen?
		cam.ViewportPosition.Y = 0
	}

	bounds := component.Sprite.Get(entry).Image.Bounds()
	levelWidth := float64(bounds.Dx())

	screenWidth := float64(game.Dimensions.ScreenWidth)
	screenWorldWidth := screenWidth / cam.ViewportZoom
	viewportWorldWidth := float64(cam.Viewport.Bounds().Dx()) / cam.ViewportZoom

	if character == nil {
		targetPos := math.Vec2{
			X: levelWidth / 2.0,
			Y: cam.ViewportPosition.Y,
		}

		target := NewTagged(w, "ViewportTarget").
			WithParent(entry).
			WithPosition(targetPos).
			With(component.Velocity).
			WithBounds(engine.Size{
				Width:  50,
				Height: 50,
			}).
			Entry()

		component.Velocity.Get(target).Velocity = math.Vec2{
			X: 1,
		}

		cam.ViewportTarget = target

		cam.ViewportBounds.X = &engine.FloatRange{
			Min: float64(-scrollingLevelCameraMargin(w)),
			Max: levelWidth + float64(scrollingLevelCameraMargin(w)) - viewportWorldWidth,
		}
	} else {
		cam.ViewportPosition.X = levelWidth/2.0 - screenWorldWidth/2.0
		cam.ViewportTarget = character

		cam.ViewportBounds.X = &engine.FloatRange{
			Min: float64(-levelCameraMargin(w)),
			Max: levelWidth + float64(levelCameraMargin(w)) - viewportWorldWidth,
		}
	}
}

func ChangeLevel(w donburi.World, level domain.TargetLevel) {
	currentLevel, ok := engine.FindWithComponent(w, component.Level)
	if ok {
		lvl := component.Level.Get(currentLevel)
		lvl.Changing = true
		game := component.MustFindGame(w)

		character, characterFound := engine.FindWithComponent(w, component.Character)
		if characterFound {
			var characterPos *domain.CharacterPosition
			pos := transform.GetTransform(character).LocalPosition
			flipY := component.Sprite.Get(character).FlipY
			characterPos = &domain.CharacterPosition{
				LocalPosition: pos,
				FlipY:         flipY,
			}

			game.PreviousLevel = &component.PreviousLevel{
				Name:              lvl.Name,
				CharacterPosition: characterPos,
			}
		}

		anim := component.Animator.Get(currentLevel)
		anim.Start("transition", currentLevel)
		transition := anim.Animations["transition"]
		transition.OnStop = func(e *donburi.Entry) {
			transform.RemoveRecursive(e)
			NewLevel(w, level)
		}
		anim.SetAnimation("transition", transition)
		return
	}

	NewLevel(w, level)
}

func levelCameraMargin(w donburi.World) int {
	game := component.MustFindGame(w)
	return int(float64(game.Dimensions.ScreenWidth) * levelCameraMarginPercent)
}

func scrollingLevelCameraMargin(w donburi.World) int {
	game := component.MustFindGame(w)
	return int(float64(game.Dimensions.ScreenWidth) * scrollingLevelCameraMarginPercent)
}

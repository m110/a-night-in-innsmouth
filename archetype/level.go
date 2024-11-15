package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

const (
	levelMovementMargin = 100
	levelCameraMargin   = 200
)

func NewLevel(w donburi.World, targetLevel domain.TargetLevel) {
	level, ok := assets.Levels[targetLevel.Name]
	if !ok {
		panic("Name not found: " + targetLevel.Name)
	}

	entry := NewTagged(w, "Level").
		WithLayer(domain.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: level.Background,
		}).
		With(component.Level).
		With(component.Animator).
		Entry()

	spawned := false

	levelCam := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCam)
	input := engine.MustFindComponent[component.InputData](w, component.Input)

	anim := component.Animator.Get(entry)

	anim.AddAnimation("transition", &component.Animation{
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
	if len(level.Entrypoints) > 0 && targetLevel.Entrypoint != nil {
		entrypoint := level.Entrypoints[*targetLevel.Entrypoint]

		bounds := component.MovementBoundsData{
			Range: engine.FloatRange{
				Min: levelMovementMargin,
				Max: float64(level.Background.Bounds().Dx() - levelMovementMargin),
			},
		}

		character = NewCharacter(entry, bounds)

		transform.GetTransform(character).LocalPosition = entrypoint.Position
		component.Sprite.Get(character).FlipY = entrypoint.FlipY
	}

	cam.Root = entry

	// TODO Review the default
	if level.CameraZoom != 0 {
		cam.ViewportZoom = level.CameraZoom
	} else {
		cam.ViewportZoom = 0.4
	}

	heightDiff := float64(game.Settings.ScreenHeight) - float64(level.Background.Bounds().Dy())*cam.ViewportZoom
	if heightDiff > 0 {
		// TODO Seems misaligned anyway, review
		cam.ViewportPosition.Y = -heightDiff / 2
	} else {
		// Should not happen?
		cam.ViewportPosition.Y = 0
	}

	bounds := component.Sprite.Get(entry).Image.Bounds()
	levelWidth := float64(bounds.Dx())

	viewportWorldWidth := float64(game.Settings.ScreenWidth) / cam.ViewportZoom

	if character == nil {
		// Show the "board" on the left of the dialog
		// TODO: make this more explicit than "level without character"
		cam.ViewportPosition.X = levelWidth/2.0 - viewportWorldWidth/3.0
		cam.ViewportTarget = nil
	} else {
		cam.ViewportPosition.X = levelWidth/2.0 - viewportWorldWidth/2.0
		cam.ViewportTarget = character
	}

	viewportWidth := float64(cam.Viewport.Bounds().Dx()) / cam.ViewportZoom

	cam.ViewportBounds.X = &engine.FloatRange{
		Min: -levelCameraMargin,
		Max: levelWidth + levelCameraMargin - viewportWidth,
	}
}

func ChangeLevel(w donburi.World, level domain.TargetLevel) {
	currentLevel, ok := engine.FindWithComponent(w, component.Level)
	if ok {
		anim := component.Animator.Get(currentLevel)
		anim.Start("transition", currentLevel)
		transition := anim.Animations["transition"]
		transition.OnStop = func(e *donburi.Entry) {
			transform.RemoveRecursive(e)
			NewLevel(w, level)
		}
		anim.AddAnimation("transition", transition)
		return
	}

	NewLevel(w, level)
}

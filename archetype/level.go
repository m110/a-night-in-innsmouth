package archetype

import (
	"time"

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
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: level.Background,
		}).
		With(component.Level).
		With(component.Animation).
		Entry()

	spawned := false

	levelCam := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCam)
	input := engine.MustFindComponent[component.InputData](w, component.Input)

	component.Animation.SetValue(entry, component.AnimationData{
		Active: true,
		Timer:  engine.NewTimer(500 * time.Millisecond),
		OnStart: func(e *donburi.Entry) {
			input.Disabled = true
		},
		Update: func(e *donburi.Entry) {
			anim := component.Animation.Get(e)
			if spawned {
				cam.TransitionAlpha = anim.Timer.PercentDone()
				if anim.Timer.IsReady() {
					anim.Stop(entry)
				}
			} else {
				cam.TransitionAlpha = 1 - anim.Timer.PercentDone()
				if anim.Timer.IsReady() {
					spawned = true
					input.Disabled = false
					anim.Stop(entry)
				}
			}
		},
	})

	for _, poi := range level.POIs {
		NewPOI(entry, poi)
	}

	game := component.MustFindGame(w)

	if level.StartPassage != "" {
		ShowPassage(w, game.Story.PassageByTitle(level.StartPassage))
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
		cam.ViewportPosition.Y = -heightDiff / 2
	} else {
		// Should not happen?
		cam.ViewportPosition.Y = 0
	}

	bounds := component.Sprite.Get(entry).Image.Bounds()
	levelWidth := float64(bounds.Dx())

	viewportWorldWidth := float64(game.Settings.ScreenWidth) / cam.ViewportZoom
	cam.ViewportPosition.X = levelWidth/2.0 - viewportWorldWidth/2.0

	if character == nil {
		cam.ViewportTarget = nil
	} else {
		cam.ViewportTarget = character
	}

	cam.ViewportBounds = &engine.FloatRange{
		Min: -levelCameraMargin,
		Max: levelWidth + levelCameraMargin,
	}
}

func ChangeLevel(w donburi.World, level domain.TargetLevel) {
	currentLevel, ok := engine.FindWithComponent(w, component.Level)
	if ok {
		anim := component.Animation.Get(currentLevel)
		anim.Start(currentLevel)
		anim.OnStop = func(e *donburi.Entry) {
			transform.RemoveRecursive(e)
			NewLevel(w, level)
		}
		return
	}

	NewLevel(w, level)
}

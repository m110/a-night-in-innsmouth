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
	levelMovementMargin = 100
)

func NewLevel(w donburi.World, targetLevel domain.TargetLevel) (*donburi.Entry, *donburi.Entry) {
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
		Entry()

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
			MinX: levelMovementMargin,
			MaxX: float64(level.Background.Bounds().Dx() - levelMovementMargin),
		}

		character = NewCharacter(entry, bounds)

		transform.GetTransform(character).LocalPosition = entrypoint.Position
		component.Sprite.Get(character).FlipY = entrypoint.FlipY
	}

	return entry, character
}

func ChangeLevel(w donburi.World, level domain.TargetLevel) {
	currentLevel := engine.MustFindWithComponent(w, component.Level)
	transform.RemoveRecursive(currentLevel)

	newLevel, character := NewLevel(w, level)

	levelCam := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCam)
	cam.Root = newLevel
	if character == nil {
		bounds := component.Sprite.Get(newLevel).Image.Bounds()
		game := component.MustFindGame(w)
		cam.ViewportPosition = math.Vec2{
			X: float64(bounds.Dx()/2.0) - float64(game.Settings.ScreenWidth/2),
			// TODO Should this be hardcoded?
			Y: -80,
		}
		cam.ViewportTarget = nil
	} else {
		cam.ViewportTarget = character
	}
}

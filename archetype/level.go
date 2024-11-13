package archetype

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewLevel(w donburi.World, levelName string) *donburi.Entry {
	level, ok := assets.Levels[levelName]
	if !ok {
		panic("Level not found: " + levelName)
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

	return entry
}

func ChangeLevel(w donburi.World, level string) {
	currentLevel := engine.MustFindWithComponent(w, component.Level)
	component.Destroy(currentLevel)
	newLevel := NewLevel(w, level)

	levelCam := engine.MustFindWithComponent(w, component.LevelCamera)
	component.Camera.Get(levelCam).Root = newLevel

	// TODO Entry points
	// TODO Should hide the character if no entry point
	character := engine.MustFindWithComponent(w, component.Character)
	transform.ChangeParent(character, newLevel, false)
}

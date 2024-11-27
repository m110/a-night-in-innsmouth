package domain

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/engine"
)

type Assets struct {
	Story     RawStory
	Levels    map[string]Level
	Character Character
	Sounds    Sounds
	Music     map[string][]byte

	TitleBackground *ebiten.Image
}

type Sounds struct {
	Click1 []byte
}

type Character struct {
	Frames   []*ebiten.Image
	Collider engine.Rect
}

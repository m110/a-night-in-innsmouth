package domain

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/engine"
)

type Assets struct {
	Story RawStory

	Settings Settings

	Levels    map[string]Level
	Character Character
	Sounds    Sounds
	Music     map[string][]byte

	// TODO Move out
	NightOverlay *ebiten.Image

	TitleBackground *ebiten.Image
}

type Settings struct {
	Character SettingsCharacter `toml:"character"`
}

type SettingsCharacter struct {
	MoveSpeed  float64  `toml:"moveSpeed"`
	StartMoney int      `toml:"startMoney"`
	StartItems []string `toml:"startItems"`
}

type Sounds struct {
	Click1 []byte
}

type Character struct {
	Frames   []*ebiten.Image
	Collider engine.Rect
}

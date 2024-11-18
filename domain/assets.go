package domain

import "github.com/hajimehoshi/ebiten/v2"

type Assets struct {
	Story     RawStory
	Levels    map[string]Level
	Character []*ebiten.Image
}

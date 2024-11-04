package component

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type SpritePivot int

const (
	SpritePivotTopLeft SpritePivot = iota
	SpritePivotCenter
)

type SpriteData struct {
	Image *ebiten.Image
	Pivot SpritePivot

	// The original rotation of the sprite
	// "Facing right" is considered 0 degrees
	OriginalRotation float64

	Hidden bool

	ColorOverride *ColorOverride
	AlphaOverride *AlphaOverride
}

type ColorOverride struct {
	R, G, B float64
}

type AlphaOverride struct {
	A float64
}

func (s *SpriteData) Show() {
	s.Hidden = false
}

func (s *SpriteData) Hide() {
	s.Hidden = true
}

var Sprite = donburi.NewComponentType[SpriteData]()

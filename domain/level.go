package domain

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/engine"
)

type Level struct {
	Name        string
	Background  func() *ebiten.Image
	POIs        []POI
	Objects     []Object
	Entrypoints []Entrypoint
	CameraZoom  float64
	Character   LevelCharacter
	Limits      *engine.FloatRange
	Fadepoint   *math.Vec2
	Outdoor     bool
}

type LevelCharacter struct {
	PosY  float64
	Scale float64
}

type Entrypoint struct {
	Index             int
	CharacterPosition CharacterPosition
}

type POI struct {
	Object       Object
	ID           string
	TriggerRect  engine.Rect
	Passage      string
	Level        *TargetLevel
	EdgeTrigger  *Direction
	TouchTrigger bool
}

type Object struct {
	Image    *ebiten.Image
	Position math.Vec2
	Scale    math.Vec2
	Layer    LayerID
}

type Direction string

var (
	EdgeLeft  Direction = "left"
	EdgeRight Direction = "right"
)

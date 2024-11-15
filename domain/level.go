package domain

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/engine"
)

type Level struct {
	Background   *ebiten.Image
	POIs         []POI
	Objects      []Object
	StartPassage string
	Entrypoints  []Entrypoint
	CameraZoom   float64
}

type Entrypoint struct {
	Index             int
	CharacterPosition CharacterPosition
}

type POI struct {
	ID          string
	Image       *ebiten.Image
	TriggerRect engine.Rect
	Rect        engine.Rect
	Passage     string
	Level       *TargetLevel
	EdgeTrigger *Edge
}

type Object struct {
	Image    *ebiten.Image
	Position math.Vec2
	Scale    math.Vec2
	Layer    LayerID
}

type Edge string

var (
	EdgeLeft  Edge = "left"
	EdgeRight Edge = "right"
)

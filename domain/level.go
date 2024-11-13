package domain

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/engine"
)

type Level struct {
	Background   *ebiten.Image
	POIs         []POI
	StartPassage string
	Entrypoints  []Entrypoint
}

type Entrypoint struct {
	Index    int
	Position math.Vec2
	FlipY    bool
}

type POI struct {
	ID          string
	TriggerRect engine.Rect
	Rect        engine.Rect
	Passage     string
	Level       *TargetLevel
}

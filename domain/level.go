package domain

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/engine"
)

type Level struct {
	Background *ebiten.Image
	POIs       []POI
}

type POI struct {
	Rect    engine.Rect
	Passage string
	Level   string
}

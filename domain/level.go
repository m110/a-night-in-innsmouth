package domain

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/engine"
)

type Level struct {
	Background   *ebiten.Image
	POIs         []POI
	StartPassage string
}

type POI struct {
	ID          string
	TriggerRect engine.Rect
	Rect        engine.Rect
	Passage     string
	Level       string
}

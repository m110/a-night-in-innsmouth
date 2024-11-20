package component

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/engine"
)

type TextSize int

const (
	TextSizeM TextSize = iota
	TextSizeS
	TextSizeL
)

type TextData struct {
	Text  string
	Color color.Color
	Size  TextSize

	Align text.Align

	Hidden bool

	Streaming      bool
	StreamingTimer *engine.Timer
}

var Text = donburi.NewComponentType[TextData]()

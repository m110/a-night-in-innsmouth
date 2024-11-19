package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Title struct {
	screenWidth     int
	screenHeight    int
	newGameCallback func()
}

func NewTitle(screenWidth int, screenHeight int, newGameCallback func()) *Title {
	return &Title{
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		newGameCallback: newGameCallback,
	}
}

func (t *Title) Update() {
	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
		t.newGameCallback()
		return
	}

	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 {
		t.newGameCallback()
		return
	}
}

func (t *Title) Draw(screen *ebiten.Image) {
}

func (t *Title) OnLayoutChange(width, height int) {
	// TODO implement me
}

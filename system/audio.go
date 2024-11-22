package system

import (
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/domain"
)

type Audio struct {
	audioContext *audio.Context

	click1Player *audio.Player
}

func NewAudio() *Audio {
	ctx := audio.CurrentContext()

	return &Audio{
		audioContext: ctx,

		click1Player: ctx.NewPlayerFromBytes(assets.Assets.Sounds.Click1),
	}
}

func (a *Audio) Init(w donburi.World) {
	domain.ButtonClickedEvent.Subscribe(w, a.onButtonClicked)
}

func (a *Audio) Update(w donburi.World) {}

func (a *Audio) onButtonClicked(w donburi.World, event domain.ButtonClicked) {
	_ = a.click1Player.Rewind()
	a.click1Player.Play()
}

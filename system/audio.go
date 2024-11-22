package system

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

type Audio struct {
	audioContext *audio.Context

	click1Player *audio.Player

	currentTrack       string
	currentMusicPlayer *audio.Player
	fadingMusicPlayer  *audio.Player
	musicMixTimer      *engine.Timer
	mixInProgress      bool
}

func NewAudio() *Audio {
	ctx := audio.CurrentContext()

	return &Audio{
		audioContext: ctx,

		click1Player: ctx.NewPlayerFromBytes(assets.Assets.Sounds.Click1),

		musicMixTimer: engine.NewTimer(3 * time.Second),
	}
}

func (a *Audio) Init(w donburi.World) {
	domain.ButtonClickedEvent.Subscribe(w, a.onButtonClicked)
	domain.MusicChangedEvent.Subscribe(w, a.onMusicChanged)
}

func (a *Audio) Update(w donburi.World) {
	if a.mixInProgress {
		a.musicMixTimer.Update()
		if a.musicMixTimer.IsReady() {
			if a.fadingMusicPlayer != nil {
				a.fadingMusicPlayer.Pause()
				_ = a.fadingMusicPlayer.Close()
				a.fadingMusicPlayer = nil
			}

			a.mixInProgress = false
		} else {
			value := engine.EaseInOut(a.musicMixTimer.PercentDone())
			if a.fadingMusicPlayer != nil {
				a.fadingMusicPlayer.SetVolume(1 - value)
			}

			if a.currentMusicPlayer != nil {
				a.currentMusicPlayer.SetVolume(value)
			}
		}
	} else {
		// Loop current music
		if a.currentMusicPlayer != nil {
			if !a.currentMusicPlayer.IsPlaying() {
				_ = a.currentMusicPlayer.Rewind()
				a.currentMusicPlayer.Play()
			}
		}
	}
}

func (a *Audio) onButtonClicked(w donburi.World, event domain.ButtonClicked) {
	_ = a.click1Player.Rewind()
	a.click1Player.Play()
}

func (a *Audio) onMusicChanged(w donburi.World, event domain.MusicChanged) {
	// Do nothing, the track didn't change
	if a.currentTrack == event.Track {
		return
	}

	a.mixInProgress = true
	a.musicMixTimer.Reset()

	if a.currentMusicPlayer != nil {
		a.fadingMusicPlayer = a.currentMusicPlayer
	}

	a.currentTrack = event.Track

	if event.Track == "" {
		// Stop the music
		a.currentMusicPlayer = nil
	} else {
		a.currentMusicPlayer = a.audioContext.NewPlayerFromBytes(assets.Assets.Music[event.Track])
		a.currentMusicPlayer.SetVolume(0)
		a.currentMusicPlayer.Play()
	}
}

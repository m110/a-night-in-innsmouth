package archetype

import (
	"image/color"
	"time"

	text2 "github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

const (
	dialogWidth = 500
)

func NewDialog(
	w donburi.World,
	passage *component.Passage,
) *donburi.Entry {
	game := component.MustFindGame(w)
	pos := math.Vec2{
		X: float64(game.Settings.ScreenWidth) - dialogWidth - 25,
		Y: 0,
	}

	height := game.Settings.ScreenHeight

	backgroundImage := ebiten.NewImage(dialogWidth, height)
	backgroundImage.Fill(assets.UIBackgroundColor)

	dialog := New(w).
		WithParent(MustFindUIRoot(w)).
		WithPosition(pos).
		WithLayer(component.SpriteUILayerUI).
		With(component.Dialog).
		WithSprite(component.SpriteData{
			Image: backgroundImage,
		}).
		Entry()

	component.Dialog.SetValue(dialog, component.DialogData{
		Passage:      passage,
		ActiveOption: 0,
	})

	New(w).
		WithParent(dialog).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 220,
			Y: 20,
		}).
		WithText(component.TextData{
			Text:  passage.Title,
			Align: text2.AlignCenter,
		})

	textBg := New(w).
		WithParent(dialog).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 50,
			Y: 50,
		}).
		Entry()

	text := New(w).
		WithText(component.TextData{
			Text:           passage.Content(),
			Streaming:      true,
			StreamingTimer: engine.NewTimer(500 * time.Millisecond),
		}).
		WithParent(textBg).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 10,
			Y: 50,
		})

	AdjustTextWidth(text.Entry(), 380)

	optionImg := ebiten.NewImage(400, 32)
	optionColor := color.RGBA{
		R: 50,
		G: 50,
		B: 50,
		A: 150,
	}
	optionImg.Fill(optionColor)

	for i, link := range passage.Links() {
		op := New(w).
			WithParent(dialog).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 50,
				Y: 300 + float64(i)*60,
			}).
			WithSprite(component.SpriteData{
				Image: optionImg,
			}).
			With(component.DialogOption).
			Entry()

		if i == 0 {
			indicatorImg := ebiten.NewImage(10, 32)
			indicatorImg.Fill(colornames.Lightyellow)

			New(w).
				WithParent(op).
				WithLayerInherit().
				WithPosition(math.Vec2{
					X: -20,
					Y: 0,
				}).
				WithSprite(component.SpriteData{
					Image: indicatorImg,
				}).
				With(component.ActiveOptionIndicator)
		}

		color := assets.TextColor
		if link.AllVisited() {
			color = assets.TextDarkColor
		}

		New(w).
			WithParent(op).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 10,
				Y: 2,
			}).
			WithText(component.TextData{
				Text:  link.Text,
				Color: color,
			})
	}

	return dialog
}

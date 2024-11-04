package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/assets"

	"github.com/m110/secrets/component"
)

func NewDialog(w donburi.World, passage *component.Passage) *donburi.Entry {
	img := ebiten.NewImage(500, 400)
	img.Fill(colornames.Darkgreen)

	dialog := New(w).
		WithParent(MustFindUIRoot(w)).
		WithPosition(math.Vec2{
			X: 100,
			Y: 50,
		}).
		WithLayer(component.SpriteUILayerUI).
		With(component.Dialog).
		WithSprite(component.SpriteData{
			Image: img,
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
			Text: passage.Title,
		})

	textImg := ebiten.NewImage(400, 10)
	textImg.Fill(colornames.Darkred)

	New(w).
		WithText(component.TextData{
			Text: passage.Content,
		}).
		WithSprite(component.SpriteData{
			Image: textImg,
		}).
		WithParent(dialog).
		WithPosition(math.Vec2{
			X: 50,
			Y: 50,
		}).
		WithLayerInherit()

	optionImg := ebiten.NewImage(400, 10)
	optionImg.Fill(colornames.Darkblue)

	for i, link := range passage.Links() {
		op := New(w).
			WithParent(dialog).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 50,
				Y: 200 + float64(i)*70,
			}).
			WithSprite(component.SpriteData{
				Image: optionImg,
			}).
			With(component.DialogOption).
			Entry()

		if i == 0 {
			indicatorImg := ebiten.NewImage(10, 10)
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
		if link.Target.Visited {
			color = assets.TextDarkColor
		}

		New(w).
			WithParent(op).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 10,
				Y: 0,
			}).
			WithText(component.TextData{
				Text:  link.Text,
				Color: color,
			})
	}

	return dialog
}

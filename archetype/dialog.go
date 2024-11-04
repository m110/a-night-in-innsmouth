package archetype

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/component"
)

func NewDialog(w donburi.World, dialogData component.DialogData) *donburi.Entry {
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

	component.Dialog.SetValue(dialog, dialogData)

	textImg := ebiten.NewImage(200, 100)
	textImg.Fill(colornames.Darkred)

	New(w).
		WithText(component.TextData{
			Text: dialogData.Text,
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

	optionImg := ebiten.NewImage(200, 30)
	optionImg.Fill(colornames.Darkblue)

	for i, option := range dialogData.Options {
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
			Entry()

		New(w).
			WithParent(op).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 10,
				Y: 0,
			}).
			WithText(component.TextData{
				Text: option.Text,
			})
	}

	return dialog
}

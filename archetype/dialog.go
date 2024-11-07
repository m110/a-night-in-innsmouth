package archetype

import (
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

const (
	dialogWidth = 500
)

func NewDialog(w donburi.World) *donburi.Entry {
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
		With(component.Active).
		With(component.Dialog).
		Entry()

	component.Active.Get(dialog).Active = true

	New(w).
		WithParent(dialog).
		WithLayerInherit().
		WithSprite(component.SpriteData{
			Image: backgroundImage,
		})

	stackOffset := New(w).
		WithParent(dialog).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 0,
			Y: 150,
		}).entry

	stack := New(w).
		WithParent(stackOffset).
		WithLayerInherit().
		With(component.StackedView).
		With(component.Animation).
		Entry()

	component.Animation.SetValue(stack, component.AnimationData{
		Timer: engine.NewTimer(500 * time.Millisecond),
	})

	return dialog
}

func NextPassage(w donburi.World) *donburi.Entry {
	activePassage := engine.MustFindWithComponent(w, component.Passage)
	passage := component.Passage.Get(activePassage)

	link := passage.Passage.Links()[passage.ActiveOption]
	link.Visit()

	activePassage.RemoveComponent(component.Passage)

	dialog := engine.MustFindWithComponent(w, component.Dialog)
	stack := engine.MustFindGrandchildWithComponent(dialog, component.StackedView)
	stackedView := component.StackedView.Get(stack)

	height := passage.Height

	options := engine.FindChildrenWithComponent(activePassage, component.DialogOption)
	for _, option := range options {
		component.Destroy(option)

		opt := component.DialogOption.Get(option)
		if passage.ActiveOption == opt.Index {
			txt := engine.MustFindChildWithComponent(option, component.Text)
			transform.ChangeParent(txt, activePassage, false)
			transform.GetTransform(txt).LocalPosition.Y = height
			height += float64(opt.Lines * passageLineHeight)
		}
	}

	for _, txt := range engine.FindChildrenWithComponent(activePassage, component.Text) {
		component.Text.Get(txt).Color = assets.TextDarkColor
	}

	stackedView.CurrentY += height
	stackTransform := transform.GetTransform(stack)
	startY := stackTransform.LocalPosition.Y

	anim := component.Animation.Get(stack)
	anim.Update = func(e *donburi.Entry) {
		stackTransform.LocalPosition.Y = startY - height*anim.Timer.PercentDone()
		if anim.Timer.IsReady() {
			anim.Stop()
		}
	}
	anim.Start()

	return NewPassage(w, link.Target)
}

const (
	passageMarginLeft = 20
	passageMarginTop  = 20

	passageLineHeight = 36
)

func NewPassage(w donburi.World, domainPassage *domain.Passage) *donburi.Entry {
	dialog := engine.MustFindWithComponent(w, component.Dialog)
	stack := engine.MustFindGrandchildWithComponent(dialog, component.StackedView)
	stackedView := component.StackedView.Get(stack)

	passage := New(w).
		WithParent(stack).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: passageMarginLeft,
			Y: stackedView.CurrentY + passageMarginTop,
		}).
		With(component.Passage).
		Entry()

	New(w).
		WithParent(passage).
		WithLayer(component.SpriteUILayerText).
		WithPosition(math.Vec2{
			X: 220,
			Y: 20,
		}).
		WithText(component.TextData{
			Text:  domainPassage.Title,
			Align: text.AlignCenter,
		})

	textBg := New(w).
		WithParent(passage).
		WithLayer(component.SpriteUILayerText).
		WithPosition(math.Vec2{
			X: 20,
			Y: 30,
		}).
		Entry()

	txt := New(w).
		WithText(component.TextData{
			Text:           domainPassage.Content(),
			Streaming:      true,
			StreamingTimer: engine.NewTimer(500 * time.Millisecond),
		}).
		WithParent(textBg).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 10,
			Y: 20,
		})

	passageLines := 1

	adjusted := AdjustTextWidth(txt.Entry(), 380)
	passageLines += strings.Count(adjusted, "\n") + 1

	optionColor := color.RGBA{
		R: 50,
		G: 50,
		B: 50,
		A: 150,
	}

	optionImageWidth := 400
	optionWidth := 380
	currentY := 300
	heightPerLine := 28
	paddingPerLine := 4

	for i, link := range domainPassage.Links() {
		op := New(w).
			WithParent(passage).
			WithLayer(component.SpriteUILayerButtons).
			WithSprite(component.SpriteData{}).
			With(component.Collider).
			With(component.DialogOption).
			Entry()

		if i == 0 {
			indicatorImg := ebiten.NewImage(10, heightPerLine+paddingPerLine)
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

		opText := New(w).
			WithParent(op).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 10,
				Y: 2,
			}).
			WithText(component.TextData{
				Text:  link.Text,
				Color: color,
			}).
			Entry()

		newText := AdjustTextWidth(opText, optionWidth)
		lines := strings.Count(newText, "\n") + 1

		component.DialogOption.SetValue(op, component.DialogOptionData{
			Index: i,
			Lines: lines,
		})

		lineHeight := heightPerLine*lines + paddingPerLine
		optionImg := ebiten.NewImage(optionImageWidth, lineHeight)
		optionImg.Fill(optionColor)

		transform.GetTransform(op).LocalPosition = math.Vec2{
			X: 50,
			Y: float64(currentY),
		}
		component.Sprite.Get(op).Image = optionImg
		component.Collider.SetValue(op, component.ColliderData{
			Width:  float64(optionImageWidth),
			Height: float64(lineHeight),
			Layer:  component.CollisionLayerButton,
		})

		currentY += lineHeight + 24
	}

	component.Passage.SetValue(passage, component.PassageData{
		Passage:      domainPassage,
		ActiveOption: 0,
		Height:       float64(passageLines * passageLineHeight),
	})

	return passage
}

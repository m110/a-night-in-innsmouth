package archetype

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

const (
	dialogWidth = 500

	passageMargin = 32
)

func NewDialog(w donburi.World, parent *donburi.Entry) *donburi.Entry {
	game := component.MustFindGame(w)
	pos := math.Vec2{
		X: float64(game.Settings.ScreenWidth) - dialogWidth - 25,
		Y: 0,
	}

	height := game.Settings.ScreenHeight

	backgroundImage := ebiten.NewImage(dialogWidth, height)
	backgroundImage.Fill(assets.UIBackgroundColor)

	dialog := NewTagged(w, "Dialog").
		WithParent(parent).
		WithPosition(pos).
		WithLayer(component.SpriteUILayerUI).
		With(component.Active).
		With(component.Dialog).
		Entry()

	component.Active.Get(dialog).Active = true

	NewTagged(w, "Dialog Background").
		WithParent(dialog).
		WithLayer(component.SpriteUILayerBackground).
		WithSprite(component.SpriteData{
			Image: backgroundImage,
		})

	return dialog
}

func NewDialogLog(w donburi.World) *donburi.Entry {
	game := component.MustFindGame(w)
	// TODO deduplicate
	pos := math.Vec2{
		X: float64(game.Settings.ScreenWidth) - dialogWidth - 25,
		Y: 0,
	}

	height := game.Settings.ScreenHeight
	log := NewTagged(w, "Log").
		WithLayer(component.SpriteUILayerUI).
		With(component.DialogLog).
		With(component.StackedView).
		With(component.Animation).
		Entry()

	cameraHeight := height - 300
	// TODO not sure if best place for this
	cam := NewCamera(
		w,
		pos,
		engine.Size{Width: dialogWidth, Height: cameraHeight},
		2,
		log,
	)

	component.Camera.Get(cam).Mask = CreateScrollMask(dialogWidth, cameraHeight)

	component.Animation.SetValue(log, component.AnimationData{
		Timer: engine.NewTimer(500 * time.Millisecond),
	})

	return log
}

func CreateScrollMask(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)

	fadeHeight := 50

	for y := 0; y < height; y++ {
		var alpha uint8 = 255

		if y < fadeHeight {
			alpha = uint8(float64(y) / float64(fadeHeight) * 255)
		} else if y > height-fadeHeight {
			distFromBottom := height - y
			alpha = uint8(float64(distFromBottom) / float64(fadeHeight) * 255)
		}

		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{A: alpha})
		}
	}

	return img
}

func NextPassage(w donburi.World) *donburi.Entry {
	activePassage := engine.MustFindWithComponent(w, component.Passage)
	passage := component.Passage.Get(activePassage)

	link := passage.Passage.Links()[passage.ActiveOption]
	link.Visit()

	activePassage.RemoveComponent(component.Passage)

	dialogLog := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(dialogLog)

	height := passage.Height

	for _, txt := range engine.FindChildrenWithComponent(activePassage, component.Text) {
		component.Text.Get(txt).Color = assets.TextDarkColor
	}

	q := donburi.NewQuery(filter.And(filter.Contains(component.DialogOption)))
	var options []*donburi.Entry
	q.Each(w, func(e *donburi.Entry) {
		options = append(options, e)
	})

	for _, e := range options {
		opt := component.DialogOption.Get(e)
		if passage.ActiveOption == opt.Index {
			txt := engine.MustFindChildWithComponent(e, component.Text)

			t := component.Text.Get(txt)

			newOption := NewTagged(w, "Option Selected").
				WithParent(activePassage).
				WithLayerInherit().
				WithPosition(math.Vec2{
					X: 2,
					Y: height + passageMargin,
				}).
				WithText(component.TextData{
					Text:  fmt.Sprintf("-> %s", t.Text),
					Color: assets.TextBlueColor,
				}).
				With(component.Bounds).
				Entry()

			AdjustTextWidth(newOption, passageTextWidth)

			textHeight := MeasureTextHeight(newOption)
			height += passageMargin + textHeight
			component.Bounds.SetValue(newOption, component.BoundsData{
				Width:  passageTextWidth,
				Height: textHeight,
			})
		}

		component.Destroy(e)
	}

	stackedView.CurrentY += height
	stackTransform := transform.GetTransform(dialogLog)
	startY := stackTransform.LocalPosition.Y

	anim := component.Animation.Get(dialogLog)
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
	passageMarginTop  = 250

	passageTextWidth = 380
)

func NewPassage(w donburi.World, domainPassage *domain.Passage) *donburi.Entry {
	log := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(log)

	passage := NewTagged(w, "Passage").
		WithParent(log).
		WithLayer(component.SpriteUILayerText).
		WithPosition(math.Vec2{
			X: passageMarginLeft,
			Y: stackedView.CurrentY + passageMarginTop,
		}).
		With(component.Passage).
		Entry()

	textY := float64(passageMargin)
	passageHeight := textY

	if domainPassage.Header != "" {
		header := NewTagged(w, "Header").
			WithParent(passage).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 220,
				Y: 20,
			}).
			WithText(component.TextData{
				Text:  domainPassage.Header,
				Align: text.AlignCenter,
			}).
			With(component.Bounds).
			Entry()

		textHeight := MeasureTextHeight(header)

		component.Bounds.SetValue(header, component.BoundsData{
			Width:  passageTextWidth,
			Height: textHeight,
		})

		headerMargin := 20.0

		textY += textHeight + headerMargin
		passageHeight += textHeight + headerMargin
	}

	txt := NewTagged(w, "Passage Text").
		WithText(component.TextData{
			Text:           domainPassage.Content(),
			Streaming:      true,
			StreamingTimer: engine.NewTimer(500 * time.Millisecond),
		}).
		WithParent(passage).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 10,
			Y: textY,
		}).
		With(component.Bounds).
		Entry()

	AdjustTextWidth(txt, passageTextWidth)
	textHeight := MeasureTextHeight(txt)
	passageHeight += textHeight

	component.Bounds.SetValue(txt, component.BoundsData{
		Width:  passageTextWidth,
		Height: textHeight,
	})

	optionColor := color.RGBA{
		R: 50,
		G: 50,
		B: 50,
		A: 150,
	}

	optionImageWidth := 400
	optionWidth := 380
	currentY := 500
	heightPerLine := 28
	paddingPerLine := 4

	dialog := engine.MustFindWithComponent(w, component.Dialog)

	for i, link := range domainPassage.Links() {
		op := NewTagged(w, "Option").
			WithParent(dialog).
			WithLayer(component.SpriteUILayerButtons).
			WithSprite(component.SpriteData{}).
			With(component.Collider).
			With(component.DialogOption).
			Entry()

		if i == 0 {
			indicatorImg := ebiten.NewImage(10, heightPerLine+paddingPerLine)
			indicatorImg.Fill(colornames.Lightyellow)

			NewTagged(w, "Indicator").
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

		color := assets.TextBlueColor
		if link.AllVisited() {
			color = assets.TextDarkColor
		}

		opText := NewTagged(w, "Option Text").
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
			X: 30,
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
		Height:       passageHeight,
	})

	return passage
}

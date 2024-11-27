package archetype

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
	passageMarginTopPercent  = 0.05
	passageMarginLeftPercent = 0.05

	scrollMaskHeightPercent = 0.05

	openDialogDuration = 1000 * time.Millisecond

	defaultParagraphEffect         = domain.ParagraphEffectTyping
	defaultParagraphEffectDuration = 500 * time.Millisecond
)

func NewDialog(w donburi.World) *donburi.Entry {
	game := component.MustFindGame(w)
	dialogWidth := game.Dimensions.DialogWidth

	pos := math.Vec2{
		X: float64(game.Dimensions.ScreenWidth - dialogWidth),
		Y: 0,
	}
	screenHeight := game.Dimensions.ScreenHeight

	backgroundImage := ebiten.NewImage(dialogWidth, screenHeight)
	backgroundImage.Fill(assets.UIBackgroundColor)

	dialog := NewTagged(w, "Dialog").
		WithLayer(domain.SpriteUILayerBackground).
		WithSprite(component.SpriteData{
			Image: backgroundImage,
		}).
		With(component.Active).
		With(component.Dialog).
		Entry()

	dialogCamera := NewCamera(
		w,
		pos,
		engine.Size{Width: dialogWidth, Height: screenHeight},
		2,
		dialog,
	)
	dialogCamera.AddComponent(component.DialogCamera)
	dialogCamera.AddComponent(component.Animator)
	dialogCamera.AddComponent(component.Active)

	cam := component.Camera.Get(dialogCamera)

	input := engine.MustFindComponent[component.InputData](w, component.Input)

	anim := component.Animator.Get(dialogCamera)
	anim.SetAnimation("fade-in", &component.Animation{
		Active: false,
		Timer:  engine.NewTimer(openDialogDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			t := a.Timer.PercentDone()
			t = engine.EaseInOut(t)

			if a.Timer.IsReady() {
				a.Stop(e)
			}

			cam.AlphaOverride = &component.AlphaOverride{
				A: t,
			}
		},
		OnStart: func(e *donburi.Entry) {
			cam.AlphaOverride = &component.AlphaOverride{
				A: 0,
			}
			input.Disabled = true
		},
		OnStop: func(e *donburi.Entry) {
			input.Disabled = false
			cam.AlphaOverride = nil
		},
	})
	anim.SetAnimation("fade-out", &component.Animation{
		Active: false,
		Timer:  engine.NewTimer(openDialogDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			t := a.Timer.PercentDone()
			t = engine.EaseInOut(t)

			if a.Timer.IsReady() {
				a.Stop(e)
			}

			cam.AlphaOverride = &component.AlphaOverride{
				A: 1 - t,
			}
		},
		OnStart: func(e *donburi.Entry) {
			cam.AlphaOverride = &component.AlphaOverride{
				A: 1,
			}
			input.Disabled = true
		},
		OnStop: func(e *donburi.Entry) {
			input.Disabled = false
			cam.AlphaOverride = nil
			component.Active.Get(dialog).Active = false
		},
	})

	im := ebiten.NewImage(dialogWidth, screenHeight)
	im.Fill(colornames.Yellow)
	log := NewTagged(w, "Log").
		WithLayer(domain.SpriteUILayerTop).
		With(component.DialogLog).
		With(component.StackedView).
		Entry()

	logCameraHeight := game.Dimensions.DialogLogHeight
	logCamera := NewCamera(
		w,
		math.Vec2{},
		engine.Size{Width: dialogWidth, Height: logCameraHeight},
		2,
		log,
	)

	transform.AppendChild(dialog, logCamera, true)
	logCamera.AddComponent(component.InnerCamera)
	logCamera.AddComponent(component.DialogLogCamera)
	logCamera.AddComponent(component.Animator)

	logAnim := component.Animator.Get(logCamera)
	logAnim.SetAnimation("scroll", &component.Animation{
		Timer: engine.NewTimer(500 * time.Millisecond),
	})

	logCam := component.Camera.Get(logCamera)
	logCam.Mask = CreateScrollMask(w, dialogWidth, logCameraHeight)
	logCam.ViewportBounds.Y = &engine.FloatRange{
		Min: 0,
		Max: 0,
	}

	stackedView := component.StackedView.Get(log)
	stackedView.CurrentY = float64(logCameraHeight) - float64(engine.IntPercent(game.Dimensions.DialogLogHeight, scrollMaskHeightPercent))

	return log
}

func CreateScrollMask(w donburi.World, width, height int) *ebiten.Image {
	game := component.MustFindGame(w)
	img := ebiten.NewImage(width, height)

	scrollMaskHeight := engine.IntPercent(game.Dimensions.DialogLogHeight, scrollMaskHeightPercent)

	for y := 0; y < height; y++ {
		var alpha uint8 = 255

		if y < scrollMaskHeight {
			alpha = uint8(float64(y) / float64(scrollMaskHeight) * 255)
		} else if y > height-scrollMaskHeight {
			distFromBottom := height - y
			alpha = uint8(float64(distFromBottom) / float64(scrollMaskHeight) * 255)
		}

		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{A: alpha})
		}
	}

	return img
}

func NextPassage(w donburi.World) {
	activePassage := engine.MustFindWithComponent(w, component.Passage)
	passage := component.Passage.Get(activePassage)

	link := passage.Passage.Links()[passage.ActiveOption]
	link.Visit()

	activePassage.RemoveComponent(component.Passage)

	for entry := range donburi.NewQuery(filter.Contains(component.RecentParagraph)).Iter(w) {
		component.Text.Get(entry).Color = assets.TextDarkColor
		entry.RemoveComponent(component.RecentParagraph)
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
			paragraph := domain.Paragraph{
				Text: fmt.Sprintf("-> %s", t.Text),
				Type: domain.ParagraphTypeRead,
			}
			AddLogParagraph(w, paragraph, ParagraphOptions{})
		}

		component.Destroy(e)
	}

	levelCamera := engine.MustFindWithComponent(w, component.LevelCamera)
	briefZoom := component.BriefZoom.Get(levelCamera)

	if link.IsExit() {
		_, ok := engine.FindWithComponent(w, component.Character)
		// Character found: zoom out back on the character
		if ok {
			hideDialog(w, nil)

			// Refresh POIs in case the conditions to show the passage changed
			DeactivatePOIs(w)
			CheckNextPOI(w)

			if briefZoom.OriginCamera != nil {
				lCam := component.Camera.Get(levelCamera)

				zoomAnim := newCameraZoomAnimation(
					lCam,
					lCam.ViewportPosition,
					briefZoom.OriginCamera.ViewportPosition,
					lCam.ViewportZoom,
					briefZoom.OriginCamera.ViewportZoom,
				)

				zoomAnim.OnStop = func(e *donburi.Entry) {
					lCam.ViewportBounds = briefZoom.OriginCamera.ViewportBounds
					lCam.ViewportTarget = briefZoom.OriginCamera.ViewportTarget
				}

				component.Animator.Get(levelCamera).SetAnimation("zoom-out", zoomAnim)
			}
		} else {
			// Character not found: Go back to the previous level
			game := component.MustFindGame(w)

			if game.PreviousLevel == nil {
				panic("no character present and no previous level found")
			}

			lvl := domain.TargetLevel{
				Name:              game.PreviousLevel.Name,
				CharacterPosition: game.PreviousLevel.CharacterPosition,
			}

			hideDialog(w, func(e *donburi.Entry) {
				ChangeLevel(w, lvl)
			})
		}

		return
	}

	if link.HasTag("main-menu") {
		// TODO Fade out on all cameras
		game := component.MustFindGame(w)
		game.SwitchToTitle()
		return
	}

	if link.Level != nil {
		// When switching levels, the camera should be reset
		briefZoom.OriginCamera = nil

		hideDialog(w, func(e *donburi.Entry) {
			ChangeLevel(w, *link.Level)
		})
		return
	}

	ShowPassage(w, link.Target, nil)
}

func extendDialogLog(w donburi.World, height float64) {
	dialogLog := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(dialogLog)
	stackedView.CurrentY += height
}

func scrollDialogLog(w donburi.World) {
	dialogLog := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(dialogLog)

	cameraEntry := engine.MustFindWithComponent(w, component.DialogLogCamera)
	cam := component.Camera.Get(cameraEntry)
	camHeight := float64(cam.Viewport.Bounds().Dy())

	game := component.MustFindGame(w)

	scrollMaskHeight := engine.IntPercent(game.Dimensions.DialogLogHeight, scrollMaskHeightPercent)

	marginBottom := float64(scrollMaskHeight)
	endY := stackedView.CurrentY - camHeight + marginBottom

	if endY < 0 {
		return
	}

	cam.ViewportBounds.Y = &engine.FloatRange{
		Min: 0,
		Max: endY,
	}

	startY := cam.ViewportPosition.Y
	scrollValue := endY - startY

	anim := component.Animator.Get(cameraEntry)
	scroll := anim.Animations["scroll"]
	scroll.Update = func(e *donburi.Entry, a *component.Animation) {
		cam.ViewportPosition.Y = startY + scrollValue*engine.EaseInOut(a.Timer.PercentDone())
		if a.Timer.IsReady() {
			a.Stop(cameraEntry)
		}
	}
	scroll.Start(cameraEntry)
}

func showDialog(w donburi.World) {
	dialog := engine.MustFindWithComponent(w, component.Dialog)
	if component.Active.Get(dialog).Active {
		return
	}

	dialogCamera := engine.MustFindWithComponent(w, component.DialogCamera)
	component.Active.Get(dialog).Active = true
	component.Active.Get(dialogCamera).Active = true

	component.Animator.Get(dialogCamera).Animations["fade-in"].Start(dialog)
}

func hideDialog(w donburi.World, onHide func(e *donburi.Entry)) {
	dialog := engine.MustFindWithComponent(w, component.Dialog)
	if !component.Active.Get(dialog).Active {
		return
	}

	dialogCamera := engine.MustFindWithComponent(w, component.DialogCamera)
	anim := component.Animator.Get(dialogCamera).Animations["fade-out"]
	anim.Start(dialogCamera)

	if onHide != nil {
		anim.OnStopOneShot = append(anim.OnStopOneShot, onHide)
	}

	anim.Start(dialog)
}

func ShowPassage(w donburi.World, domainPassage *domain.Passage, source *donburi.Entry) {
	if len(domainPassage.Paragraphs) == 0 && len(domainPassage.Links()) == 0 {
		// In rare cases, a passage might have no paragraphs or links, so we skip it
		// (e.g. when a passage is only used for macros)
		return
	}

	if source != nil {
		zoomInOnPOI(w, source)
	}

	showDialog(w)

	log := engine.MustFindWithComponent(w, component.DialogLog)

	passage := NewTagged(w, "Passage").
		WithParent(log).
		WithLayer(domain.SpriteUILayerText).
		With(component.Passage).
		Entry()

	component.Passage.SetValue(passage, component.PassageData{
		Passage:      domainPassage,
		ActiveOption: 0,
	})

	paragraphs := domainPassage.AvailableParagraphs()

	if len(paragraphs) == 0 {
		createDialogOptions(w, domainPassage)
	}

	delay := time.Duration(0)
	for i, paragraph := range paragraphs {
		opt := ParagraphOptions{
			Delay: delay,
		}

		if i == len(paragraphs)-1 {
			opt.OnShow = func() {
				createDialogOptions(w, domainPassage)
			}
		}

		AddLogParagraph(w, paragraph, opt)

		delay += paragraph.Delay

		if paragraph.EffectDuration == 0 {
			delay += defaultParagraphEffectDuration
		} else {
			delay += paragraph.EffectDuration
		}
	}
}

type ParagraphOptions struct {
	Delay  time.Duration
	OnShow func()
}

func AddLogParagraph(w donburi.World, paragraph domain.Paragraph, options ParagraphOptions) {
	game := component.MustFindGame(w)

	log := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(log)

	// TODO Deduplicate
	passageMarginLeft := engine.IntPercent(game.Dimensions.DialogWidth, passageMarginLeftPercent)
	passageTextWidth := game.Dimensions.DialogWidth - passageMarginLeft*2
	logCamera := engine.MustFindWithComponent(w, component.DialogLogCamera)
	logHeight := component.Camera.Get(logCamera).Viewport.Bounds().Dy()
	passageMarginTop := float64(int(float64(logHeight) * passageMarginTopPercent))

	textColor := assets.TextColor
	textSize := component.TextSizeM
	switch paragraph.Type {
	case domain.ParagraphTypeHeader:
		textSize = component.TextSizeL
	case domain.ParagraphTypeHint:
		textColor = assets.TextOrangeColor
	case domain.ParagraphTypeFear:
		textColor = assets.TextPurpleColor
	case domain.ParagraphTypeReceived:
		textColor = assets.TextGreenColor
	case domain.ParagraphTypeLost:
		textColor = assets.TextRedColor
	case domain.ParagraphTypeRead:
		textColor = assets.TextDarkColor
	default:
	}

	txt := component.TextData{
		Text:  paragraph.Text,
		Color: textColor,
		Size:  textSize,
	}

	entry := NewTagged(w, "Log Paragraph").
		WithParent(log).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: float64(passageMarginLeft),
			Y: stackedView.CurrentY + passageMarginTop,
		}).
		WithText(txt).
		With(component.Bounds).
		With(component.Animator).
		Entry()

	if paragraph.Type == domain.ParagraphTypeStandard {
		entry.AddComponent(component.RecentParagraph)
	}

	AdjustTextWidth(entry, passageTextWidth)

	optionTextHeight := MeasureTextHeight(entry)
	height := passageMarginTop + optionTextHeight
	extendDialogLog(w, height)

	switch paragraph.Align {
	case domain.ParagraphAlignCenter:
		transform.GetTransform(entry).LocalPosition.X += (float64(passageTextWidth) - MeasureTextWidth(entry)) / 2.0
	}

	component.Bounds.SetValue(entry, component.BoundsData{
		Width:  float64(passageTextWidth),
		Height: optionTextHeight,
	})

	animator := component.Animator.Get(entry)

	delay := paragraph.Delay + options.Delay
	animator.SetAnimation("delay", &component.Animation{
		Active: true,
		Timer:  engine.NewTimer(delay),
		Update: func(e *donburi.Entry, a *component.Animation) {
			if a.Timer.IsReady() {
				animator.Animations["show"].Start(e)
				a.Stop(e)
			}
		},
	})

	effectDuration := defaultParagraphEffectDuration
	if paragraph.EffectDuration != 0 {
		effectDuration = paragraph.EffectDuration
	}

	effect := paragraph.Effect
	if effect == domain.ParagraphEffectDefault {
		effect = defaultParagraphEffect
	}

	switch effect {
	case domain.ParagraphEffectTyping:
		t := component.Text.Get(entry)
		t.Hidden = true
	case domain.ParagraphEffectFadeIn:
		panic("not implemented")
	case domain.ParagraphEffectDefault:
	default:
		panic("unknown paragraph effect")
	}

	animator.SetAnimation("show", &component.Animation{
		Active: false,
		Timer:  engine.NewTimer(effectDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			if a.Timer.IsReady() {
				a.Stop(e)
			}
		},
		OnStart: func(e *donburi.Entry) {
			switch effect {
			case domain.ParagraphEffectTyping:
				t := component.Text.Get(e)
				t.Hidden = false
				t.Streaming = true
				t.StreamingTimer = engine.NewTimer(effectDuration)
			case domain.ParagraphEffectFadeIn:
				panic("not implemented")
			case domain.ParagraphEffectDefault:
			}

			scrollDialogLog(w)
		},
		OnStop: func(e *donburi.Entry) {
			switch effect {
			case domain.ParagraphEffectTyping:
			case domain.ParagraphEffectFadeIn:
				panic("not implemented")
			case domain.ParagraphEffectDefault:
			}

			if options.OnShow != nil {
				options.OnShow()
			}
		},
	})
}

func zoomInOnPOI(w donburi.World, source *donburi.Entry) {
	levelCamera := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCamera)
	bz := component.BriefZoom.Get(levelCamera)

	camCopy := *cam
	bz.OriginCamera = &camCopy

	cam.ViewportBounds = component.ViewportBounds{}
	cam.ViewportTarget = nil

	originPosition := cam.ViewportPosition
	originZoom := cam.ViewportZoom
	targetZoom := cam.ViewportZoom * 1.5

	bounds := component.Bounds.Get(source)

	targetWorldPos := transform.WorldPosition(source)
	viewportWidth := float64(cam.Viewport.Bounds().Dx()) / targetZoom
	viewportHeight := float64(cam.Viewport.Bounds().Dy()) / targetZoom

	targetPosition := math.Vec2{
		X: targetWorldPos.X + bounds.Width/2.0 - viewportWidth/4.0,
		Y: targetWorldPos.Y + bounds.Height/2.0 - viewportHeight/2.0,
	}

	animator := component.Animator.Get(levelCamera)
	animator.SetAnimation("zoom-in", newCameraZoomAnimation(cam, originPosition, targetPosition, originZoom, targetZoom))
}

func createDialogOptions(w donburi.World, domainPassage *domain.Passage) {
	dialog := engine.MustFindWithComponent(w, component.Dialog)

	game := component.MustFindGame(w)

	optionImageWidthPercent := 1 - passageMarginLeftPercent*2
	optionWidthPercent := 0.9

	marginLeft := engine.IntPercent(game.Dimensions.DialogWidth, passageMarginLeftPercent)
	optionImageWidth := engine.IntPercent(game.Dimensions.DialogWidth, optionImageWidthPercent) - marginLeft
	optionTextWidth := engine.IntPercent(optionImageWidth, optionWidthPercent)
	textMarginLeft := float64(optionImageWidth-optionTextWidth) / 2
	indicatorWidth := int(textMarginLeft / 2)

	// One row is one space between buttons or margin
	// Two rows are one button
	rowHeight := game.Dimensions.DialogOptionsRowHeight
	buttonHeight := rowHeight * 2

	optionsParent := New(w).
		WithParent(dialog).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: float64(marginLeft),
			Y: float64(game.Dimensions.DialogLogHeight),
		}).
		Entry()

	optionsY := rowHeight

	for i, link := range domainPassage.Links() {
		op := NewTagged(w, "Option").
			WithParent(optionsParent).
			WithLayer(domain.SpriteUILayerButtons).
			WithSprite(component.SpriteData{}).
			With(component.Collider).
			With(component.DialogOption).
			Entry()

		if i == 0 {
			indicatorImg := ebiten.NewImage(indicatorWidth, buttonHeight)
			indicatorImg.Fill(colornames.Lightyellow)

			NewTagged(w, "Indicator").
				WithParent(op).
				WithLayerInherit().
				WithSprite(component.SpriteData{
					Image: indicatorImg,
				}).
				With(component.ActiveOptionIndicator)
		}

		textColor := assets.TextBlueColor
		if link.AllVisited() {
			textColor = assets.TextDarkColor
		}

		opText := NewTagged(w, "Option Text").
			WithParent(op).
			WithLayerInherit().
			WithText(component.TextData{
				Text:  link.Text,
				Color: textColor,
			}).
			Entry()

		newText := AdjustTextWidth(opText, optionTextWidth)
		lines := strings.Count(newText, "\n") + 1

		textMarginTop := (float64(buttonHeight) - assets.NormalFont.Size*float64(lines)) / 2.0
		transform.GetTransform(opText).LocalPosition = math.Vec2{
			X: textMarginLeft,
			Y: textMarginTop,
		}

		component.DialogOption.SetValue(op, component.DialogOptionData{
			Index: i,
			Lines: lines,
		})

		lineHeight := buttonHeight * lines
		optionImg := ebiten.NewImage(optionImageWidth, buttonHeight)
		optionImg.Fill(assets.OptionColor)

		transform.GetTransform(op).LocalPosition = math.Vec2{
			Y: float64(optionsY),
		}
		component.Sprite.Get(op).Image = optionImg
		component.Collider.SetValue(op, component.ColliderData{
			Rect:  engine.NewRect(0, 0, float64(optionImageWidth), float64(lineHeight)),
			Layer: domain.CollisionLayerButton,
		})

		optionsY += lineHeight + int(assets.NormalFont.Size)
	}
}

func newCameraZoomAnimation(
	cam *component.CameraData,
	originPosition, targetPosition math.Vec2,
	originZoom, targetZoom float64,
) *component.Animation {
	return &component.Animation{
		Active: true,
		Timer:  engine.NewTimer(openDialogDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			t := a.Timer.PercentDone()

			t = engine.EaseInOut(t)

			if a.Timer.IsReady() {
				// Force exact final position and zoom when animation ends
				cam.ViewportPosition = targetPosition
				cam.ViewportZoom = targetZoom
				a.Stop(e)
			} else {
				cam.ViewportPosition = engine.LerpVec2(originPosition, targetPosition, t)
				cam.ViewportZoom = engine.Lerp(originZoom, targetZoom, t)
			}
		},
	}
}

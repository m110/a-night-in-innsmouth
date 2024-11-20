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

	LevelTransitionDuration = 500 * time.Millisecond
	openDialogDuration      = 1000 * time.Millisecond
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

	for _, txt := range engine.FindChildrenWithComponent(activePassage, component.Text) {
		component.Text.Get(txt).Color = assets.TextDarkColor
	}

	q := donburi.NewQuery(filter.And(filter.Contains(component.DialogOption)))
	var options []*donburi.Entry
	q.Each(w, func(e *donburi.Entry) {
		options = append(options, e)
	})

	game := component.MustFindGame(w)

	height := 0.0
	passageMarginLeft := engine.IntPercent(game.Dimensions.DialogWidth, passageMarginLeftPercent)
	passageTextWidth := game.Dimensions.DialogWidth - passageMarginLeft*2

	logCamera := engine.MustFindWithComponent(w, component.DialogLogCamera)
	logHeight := component.Camera.Get(logCamera).Viewport.Bounds().Dy()
	passageMarginTop := float64(int(float64(logHeight) * passageMarginTopPercent))

	for _, e := range options {
		opt := component.DialogOption.Get(e)
		if passage.ActiveOption == opt.Index {
			txt := engine.MustFindChildWithComponent(e, component.Text)

			t := component.Text.Get(txt)

			newOption := NewTagged(w, "Option Selected").
				WithParent(activePassage).
				WithLayerInherit().
				WithPosition(math.Vec2{
					X: 0,
					Y: passage.Height + passageMarginTop,
				}).
				WithText(component.TextData{
					Text:  fmt.Sprintf("-> %s", t.Text),
					Color: assets.TextDarkColor,
				}).
				With(component.Bounds).
				Entry()

			AdjustTextWidth(newOption, passageTextWidth)

			optionTextHeight := MeasureTextHeight(newOption)
			height += passageMarginTop + optionTextHeight

			component.Bounds.SetValue(newOption, component.BoundsData{
				Width:  float64(passageTextWidth),
				Height: optionTextHeight,
			})
		}

		component.Destroy(e)
	}

	scrollDialogLog(w, height)

	if link.IsExit() {
		_, ok := engine.FindWithComponent(w, component.Character)
		if ok {
			// Character found: zoom out back on the character
			hideDialog(w, nil)

			// Refresh POIs in case the conditions to show the passage changed
			DeactivatePOIs(w)
			CheckNextPOI(w)

			levelCamera := engine.MustFindWithComponent(w, component.LevelCamera)
			lCam := component.Camera.Get(levelCamera)
			bz := component.BriefZoom.Get(levelCamera)

			zoomAnim := newCameraZoomAnimation(
				lCam,
				lCam.ViewportPosition,
				bz.OriginCamera.ViewportPosition,
				lCam.ViewportZoom,
				bz.OriginCamera.ViewportZoom,
			)

			zoomAnim.OnStop = func(e *donburi.Entry) {
				lCam.ViewportBounds = bz.OriginCamera.ViewportBounds
				lCam.ViewportTarget = bz.OriginCamera.ViewportTarget
			}

			component.Animator.Get(levelCamera).SetAnimation("zoom-out", zoomAnim)
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

	if link.Level != nil {
		hideDialog(w, func(e *donburi.Entry) {
			ChangeLevel(w, *link.Level)
		})
		return
	}

	ShowPassage(w, link.Target, nil)
}

func scrollDialogLog(w donburi.World, height float64) {
	dialogLog := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(dialogLog)
	stackedView.CurrentY += height

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
	if source != nil {
		zoomInOnPOI(w, source)
	}

	showDialog(w)

	log := engine.MustFindWithComponent(w, component.DialogLog)
	stackedView := component.StackedView.Get(log)

	game := component.MustFindGame(w)
	// TODO deduplicate
	passageMarginLeft := engine.IntPercent(game.Dimensions.DialogWidth, passageMarginLeftPercent)
	passageTextWidth := game.Dimensions.DialogWidth - passageMarginLeft*2

	// TODO deduplicate
	logCamera := engine.MustFindWithComponent(w, component.DialogLogCamera)
	logHeight := component.Camera.Get(logCamera).Viewport.Bounds().Dy()
	passageMarginTop := float64(int(float64(logHeight) * passageMarginTopPercent))

	passage := NewTagged(w, "Passage").
		WithParent(log).
		WithLayer(domain.SpriteUILayerText).
		WithPosition(math.Vec2{
			X: float64(passageMarginLeft),
			Y: stackedView.CurrentY,
		}).
		With(component.Passage).
		Entry()

	textY := passageMarginTop
	passageHeight := textY

	if domainPassage.Header != "" {
		header := NewTagged(w, "Header").
			WithParent(passage).
			WithLayerInherit().
			WithText(component.TextData{
				Text: fmt.Sprintf("- %v -", domainPassage.Header),
				Size: component.TextSizeL,
			}).
			With(component.Bounds).
			Entry()

		headerWidth := MeasureTextWidth(header)

		transform.GetTransform(header).LocalPosition = math.Vec2{
			X: (float64(passageTextWidth) - headerWidth) / 2.0,
			Y: passageMarginTop,
		}

		headerHeight := MeasureTextHeight(header)

		component.Bounds.SetValue(header, component.BoundsData{
			Width:  headerWidth,
			Height: headerHeight,
		})

		headerHeight += passageMarginTop

		textY += headerHeight
		passageHeight += headerHeight
	}

	streamingTime := 500 * time.Millisecond

	for i, segment := range domainPassage.AvailableSegments() {
		segmentColor := assets.TextColor
		if segment.IsHint {
			segmentColor = assets.TextOrangeColor
		}

		txt := NewTagged(w, "Passage Segment Text").
			WithText(component.TextData{
				Text:           segment.Text,
				Color:          segmentColor,
				Streaming:      i == 0,
				Hidden:         i > 0,
				StreamingTimer: engine.NewTimer(streamingTime),
			}).
			WithParent(passage).
			With(component.Animator).
			WithLayerInherit().
			WithPosition(math.Vec2{
				X: 0,
				Y: textY,
			}).
			With(component.Bounds).
			Entry()

		if i > 0 {
			component.Animator.Get(txt).SetAnimation("stream", &component.Animation{
				Active: true,
				Timer:  engine.NewTimer(streamingTime * time.Duration(i)),
				Update: func(e *donburi.Entry, a *component.Animation) {
					if a.Timer.IsReady() {
						t := component.Text.Get(e)
						t.Hidden = false
						t.Streaming = true
						a.Stop(e)
					}
				},
			})
		}

		AdjustTextWidth(txt, passageTextWidth)
		segmentTextHeight := MeasureTextHeight(txt)

		yIncrease := segmentTextHeight
		if i != len(domainPassage.AvailableSegments())-1 {
			yIncrease += assets.NormalFont.Size
		}

		textY += yIncrease
		passageHeight += yIncrease

		component.Bounds.SetValue(txt, component.BoundsData{
			Width:  float64(passageTextWidth),
			Height: segmentTextHeight,
		})
	}

	component.Passage.SetValue(passage, component.PassageData{
		Passage:      domainPassage,
		ActiveOption: 0,
		Height:       passageHeight,
	})

	scrollDialogLog(w, passageHeight)

	createDialogOptions(w, domainPassage)
}

func zoomInOnPOI(w donburi.World, source *donburi.Entry) {
	levelCamera := engine.MustFindWithComponent(w, component.LevelCamera)
	cam := component.Camera.Get(levelCamera)
	bz := component.BriefZoom.Get(levelCamera)
	bz.OriginCamera = *cam

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
			Width:  float64(optionImageWidth),
			Height: float64(lineHeight),
			Layer:  domain.CollisionLayerButton,
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

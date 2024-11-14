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

	LevelTransitionDuration = 500 * time.Millisecond
	openDialogDuration      = 1000 * time.Millisecond
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
		WithLayer(component.SpriteUILayerBackground).
		WithSprite(component.SpriteData{
			Image: backgroundImage,
		}).
		With(component.Active).
		With(component.Dialog).
		With(component.Animator).
		Entry()

	input := engine.MustFindComponent[component.InputData](w, component.Input)

	sprite := component.Sprite.Get(dialog)
	animator := component.Animator.Get(dialog)
	animator.AddAnimation("fade-in", &component.Animation{
		Active: false,
		Timer:  engine.NewTimer(openDialogDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			t := a.Timer.PercentDone()
			t = engine.EaseInOut(t)

			if a.Timer.IsReady() {
				a.Stop(e)
			}

			sprite.AlphaOverride = &component.AlphaOverride{
				A: t,
			}
		},
		OnStart: func(e *donburi.Entry) {
			sprite.AlphaOverride = &component.AlphaOverride{
				A: 0,
			}
			input.Disabled = true
		},
		OnStop: func(e *donburi.Entry) {
			sprite.AlphaOverride = nil
			input.Disabled = false
		},
	})

	animator.AddAnimation("fade-out", &component.Animation{
		Active: false,
		Timer:  engine.NewTimer(openDialogDuration),
		Update: func(e *donburi.Entry, a *component.Animation) {
			t := a.Timer.PercentDone()
			t = engine.EaseInOut(t)

			if a.Timer.IsReady() {
				a.Stop(e)
			}

			sprite.AlphaOverride = &component.AlphaOverride{
				A: 1 - t,
			}
		},
		OnStart: func(e *donburi.Entry) {
			sprite.AlphaOverride = &component.AlphaOverride{
				A: 1,
			}
			input.Disabled = true
		},
		OnStop: func(e *donburi.Entry) {
			sprite.AlphaOverride = nil
			input.Disabled = false

			component.Active.Get(dialog).Active = false
		},
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
		Entry()

	cameraHeight := height - 300
	// TODO not sure if best place for this
	camera := NewCamera(
		w,
		pos,
		engine.Size{Width: dialogWidth, Height: cameraHeight},
		2,
		log,
	)

	camera.AddComponent(component.DialogCamera)
	camera.AddComponent(component.Animator)
	camera.AddComponent(component.Active)

	cam := component.Camera.Get(camera)
	cam.Mask = CreateScrollMask(dialogWidth, cameraHeight)

	anim := component.Animator.Get(camera)
	anim.AddAnimation("scroll", &component.Animation{
		Timer: engine.NewTimer(500 * time.Millisecond),
	})
	anim.AddAnimation("fade-in", &component.Animation{
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
		},
		OnStop: func(e *donburi.Entry) {
			cam.AlphaOverride = nil
		},
	})
	anim.AddAnimation("fade-out", &component.Animation{
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
		},
		OnStop: func(e *donburi.Entry) {
			cam.AlphaOverride = nil
		},
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

func NextPassage(w donburi.World) {
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

	cameraEntry := engine.MustFindWithComponent(w, component.DialogCamera)
	cam := component.Camera.Get(cameraEntry)
	startY := cam.ViewportPosition.Y
	anim := component.Animator.Get(cameraEntry)
	scroll := anim.Animations["scroll"]
	scroll.Update = func(e *donburi.Entry, a *component.Animation) {
		cam.ViewportPosition.Y = startY + height*a.Timer.PercentDone()
		if a.Timer.IsReady() {
			a.Stop(cameraEntry)
		}
	}
	scroll.Start(cameraEntry)

	if link.IsExit() {
		hideDialog(w, nil)

		// Refresh POIs in case the conditions to show the passage changed
		HidePOIs(w)
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

		component.Animator.Get(levelCamera).AddAnimation("zoom-out", zoomAnim)

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

func showDialog(w donburi.World) {
	dialog := engine.MustFindWithComponent(w, component.Dialog)
	if component.Active.Get(dialog).Active {
		return
	}

	dialogCamera := engine.MustFindWithComponent(w, component.DialogCamera)
	component.Active.Get(dialog).Active = true
	component.Active.Get(dialogCamera).Active = true

	component.Animator.Get(dialog).Animations["fade-in"].Start(dialog)
	component.Animator.Get(dialogCamera).Animations["fade-in"].Start(dialog)
}

func hideDialog(w donburi.World, onHide func(e *donburi.Entry)) {
	dialog := engine.MustFindWithComponent(w, component.Dialog)
	if !component.Active.Get(dialog).Active {
		return
	}

	dialogCamera := engine.MustFindWithComponent(w, component.DialogCamera)
	component.Animator.Get(dialogCamera).Animations["fade-out"].Start(dialogCamera)

	anim := component.Animator.Get(dialog).Animations["fade-out"]
	if onHide != nil {
		anim.OnStopOneShot = append(anim.OnStopOneShot, onHide)
	}

	anim.Start(dialog)
}

const (
	passageMarginLeft = 20
	passageMarginTop  = 250

	passageTextWidth = 380
)

func ShowPassage(w donburi.World, domainPassage *domain.Passage, source *donburi.Entry) *donburi.Entry {
	dialog := engine.MustFindWithComponent(w, component.Dialog)

	if source != nil {
		levelCamera := engine.MustFindWithComponent(w, component.LevelCamera)
		cam := component.Camera.Get(levelCamera)
		bz := component.BriefZoom.Get(levelCamera)
		bz.OriginCamera = *cam

		cam.ViewportBounds = nil
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
		animator.AddAnimation("zoom-in", newCameraZoomAnimation(cam, originPosition, targetPosition, originZoom, targetZoom))
	}

	showDialog(w)

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

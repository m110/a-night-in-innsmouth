package scene

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	donburievents "github.com/yohamta/donburi/features/events"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
	"github.com/m110/secrets/system"
)

type Title struct {
	world donburi.World

	systems   []System
	drawables []Drawable

	screenWidth  int
	screenHeight int

	switchToGameFunc func()
}

func NewTitle(screenWidth int, screenHeight int, switchToGame func()) *Title {
	t := &Title{
		screenWidth:      screenWidth,
		screenHeight:     screenHeight,
		switchToGameFunc: switchToGame,
	}

	t.loadLevel()

	return t
}

func (t *Title) loadLevel() {
	debug := system.NewDebug(t.loadLevel)

	t.systems = []System{
		debug,
		system.NewAnimation(),
		system.NewText(),
		system.NewAudio(),
		system.NewTimeToLive(),
		system.NewDestroy(),
		system.NewDimensions(),
	}

	t.drawables = []Drawable{
		system.NewRender(),
		debug,
	}

	t.world = t.createWorld()

	t.init()
}

func (t *Title) createWorld() donburi.World {
	world := donburi.NewWorld()

	// TODO separate common things out of Game to another component
	game := world.Entry(world.Create(component.Game, component.Input))
	component.Game.SetValue(game, component.GameData{
		Dimensions: system.CalculateDimensions(t.screenWidth, t.screenHeight),
	})

	ui := archetype.NewTagged(world, "UI").
		WithLayer(domain.SpriteUILayerUI).
		Entry()

	bgImg := assets.Assets.TitleBackground

	archetype.NewTagged(world, "Title").
		WithParent(ui).
		WithLayerInherit().
		WithSprite(component.SpriteData{
			Image: bgImg,
		}).
		WithSpriteBounds()

	uiCamera := archetype.NewCamera(world, math.Vec2{X: 0, Y: 0}, engine.Size{Width: t.screenWidth, Height: t.screenHeight}, 1, ui)

	marginPercent := 0.01
	screenHeight := float64(t.screenHeight)
	bgHeight := float64(bgImg.Bounds().Dy())

	totalMarginHeight := screenHeight * marginPercent * 2
	availableHeight := screenHeight - totalMarginHeight

	cam := component.Camera.Get(uiCamera)
	cam.ViewportZoom = availableHeight / bgHeight

	overlay := ebiten.NewImage(t.screenWidth, t.screenHeight)
	overlay.Fill(color.Black)
	cam.TransitionOverlay = overlay
	cam.TransitionAlpha = 1.0

	uiCamera.AddComponent(component.Animator)
	component.Animator.Get(uiCamera).SetAnimation("fade-in", &component.Animation{
		Active: true,
		Timer:  engine.NewTimer(2 * time.Second),
		Update: func(e *donburi.Entry, a *component.Animation) {
			if a.Timer.IsReady() {
				a.Stop(e)
				return
			}

			a.Timer.Update()
			cam.TransitionAlpha = 1 - engine.EaseInOut(a.Timer.PercentDone())
		},
	})
	component.Animator.Get(uiCamera).SetAnimation("fade-out", &component.Animation{
		Timer: engine.NewTimer(2 * time.Second),
		Update: func(e *donburi.Entry, a *component.Animation) {
			if a.Timer.IsReady() {
				a.Stop(e)
				return
			}

			a.Timer.Update()
			cam.TransitionAlpha = engine.EaseInOut(a.Timer.PercentDone())
		},
		OnStop: func(e *donburi.Entry) {
			t.switchToGame()
		},
	})

	return world
}

func (t *Title) init() {
	for _, s := range t.systems {
		if init, ok := s.(Initializer); ok {
			init.Init(t.world)
		}
	}

	for _, d := range t.drawables {
		if init, ok := d.(Initializer); ok {
			init.Init(t.world)
		}
	}
}

func (t *Title) switchToGame() {
	for _, s := range t.systems {
		if stop, ok := s.(Stopper); ok {
			stop.Stop(t.world)
		}
	}

	for _, d := range t.drawables {
		if stop, ok := d.(Stopper); ok {
			stop.Stop(t.world)
		}
	}

	t.switchToGameFunc()
}

func (t *Title) startTransition() {
	uiCamera, ok := donburi.NewQuery(filter.Contains(component.Camera)).First(t.world)
	if !ok {
		panic("no camera found")
	}
	component.Animator.Get(uiCamera).Animations["fade-out"].Start(uiCamera)
}

func (t *Title) Update() {
	for _, s := range t.systems {
		s.Update(t.world)
	}

	// TODO move to a system
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		t.startTransition()
		return
	}

	// TODO move to a system
	touchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.startTransition()
		return
	}

	donburievents.ProcessAllEvents(t.world)
}

func (t *Title) Draw(screen *ebiten.Image) {
	for _, s := range t.drawables {
		s.Draw(t.world, screen)
	}
}

func (t *Title) OnLayoutChange(width, height int) {
	t.screenWidth = width
	t.screenHeight = height

	if t.world == nil {
		return
	}

	gameEntry, ok := donburi.NewQuery(filter.Contains(component.Game)).First(t.world)
	if ok {
		game := component.Game.Get(gameEntry)
		game.Dimensions.ScreenWidth = width
		game.Dimensions.ScreenHeight = height
		game.Dimensions.Updated = true
	}
}

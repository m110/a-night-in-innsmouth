package scene

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	donburievents "github.com/yohamta/donburi/features/events"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
	"github.com/m110/secrets/system"
)

type Initializer interface {
	Init(w donburi.World)
}

type System interface {
	Update(w donburi.World)
}

type Drawable interface {
	Draw(w donburi.World, screen *ebiten.Image)
}

type Game struct {
	world donburi.World

	systems   []System
	drawables []Drawable

	screenWidth  int
	screenHeight int
}

func NewGame(screenWidth int, screenHeight int) *Game {
	g := &Game{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}

	g.loadLevel()

	return g
}

func (g *Game) loadLevel() {
	g.systems = []System{
		system.NewControls(),
		system.NewInventory(),
		system.NewVelocity(),
		system.NewCameraFollow(),
		system.NewCollision(),
		system.NewAnimation(),
		system.NewHierarchyValidator(),
		system.NewDetectPOI(),
		system.NewText(),
		system.NewTimeToLive(),
		system.NewDestroy(),
		system.NewDebug(g.loadLevel),
	}

	g.drawables = []Drawable{
		system.NewRender(),
	}

	g.world = g.createWorld()

	g.init()
}

func (g *Game) createWorld() donburi.World {
	world := donburi.NewWorld()

	story := domain.NewStory(world, assets.Assets.Story)

	game := world.Entry(world.Create(component.Game, component.Input))
	component.Game.SetValue(game, component.GameData{
		Story: story,
		Settings: component.Settings{
			ScreenWidth:  g.screenWidth,
			ScreenHeight: g.screenHeight,
		},
	})
	component.Input.SetValue(game, component.InputData{
		Disabled:      false,
		MoveRightKeys: []ebiten.Key{ebiten.KeyD, ebiten.KeyRight},
		MoveLeftKeys:  []ebiten.Key{ebiten.KeyA, ebiten.KeyLeft},
		ActionKeys:    []ebiten.Key{ebiten.KeySpace},
		MoveSpeed:     6,
	})

	world.Create(component.Debug)

	ui := archetype.NewTagged(world, "UI").
		WithLayer(domain.SpriteUILayerUI).
		Entry()

	archetype.NewDialog(world)

	g.createInventory(world, ui)

	story.AddMoney(1000)

	levelCam := archetype.NewCamera(world, math.Vec2{X: 0, Y: 0}, engine.Size{Width: g.screenWidth, Height: g.screenHeight}, 0, nil)
	levelCam.AddComponent(component.LevelCamera)
	levelCam.AddComponent(component.BriefZoom)
	levelCam.AddComponent(component.Animator)
	overlay := ebiten.NewImage(g.screenWidth, g.screenHeight)
	overlay.Fill(color.Black)
	cam := component.Camera.Get(levelCam)
	cam.TransitionOverlay = overlay
	cam.TransitionAlpha = 1.0
	archetype.NewCamera(world, math.Vec2{X: 0, Y: 0}, engine.Size{Width: g.screenWidth, Height: g.screenHeight}, 1, ui)

	entrypoint := 0
	archetype.ChangeLevel(world, domain.TargetLevel{
		Name:       "hotel",
		Entrypoint: &entrypoint,
	})

	return world
}

func (g *Game) init() {
	for _, s := range g.systems {
		if init, ok := s.(Initializer); ok {
			init.Init(g.world)
		}
	}

	for _, d := range g.drawables {
		if init, ok := d.(Initializer); ok {
			init.Init(g.world)
		}
	}
}

func (g *Game) Update() {
	for _, s := range g.systems {
		s.Update(g.world)
	}

	donburievents.ProcessAllEvents(g.world)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, s := range g.drawables {
		s.Draw(g.world, screen)
	}
}

const (
	inventoryWidth  = 250
	inventoryHeight = 50
)

func (g *Game) createInventory(w donburi.World, ui *donburi.Entry) {
	inventoryButtonImg := ebiten.NewImage(inventoryWidth, inventoryHeight)
	inventoryButtonImg.Fill(assets.UIBackgroundColor)

	inventoryButton := archetype.NewTagged(w, "Inventory Button").
		WithParent(ui).
		WithLayer(domain.SpriteUILayerUI).
		With(component.Active).
		WithSprite(component.SpriteData{
			Image: inventoryButtonImg,
		}).
		With(component.Collider).
		With(component.Inventory).
		Entry()
	component.Active.Get(inventoryButton).Active = true

	component.Collider.SetValue(inventoryButton, component.ColliderData{
		Width:  float64(inventoryWidth),
		Height: float64(inventoryHeight),
		Layer:  domain.CollisionLayerButton,
	})

	archetype.NewTagged(w, "Inventory Button Text").
		WithParent(inventoryButton).
		WithLayerInherit().
		WithText(component.TextData{
			Text: "Inventory (e)",
		}).
		WithPosition(math.Vec2{
			X: 10,
			Y: 10,
		})

	inventoryImg := ebiten.NewImage(inventoryWidth, g.screenHeight)
	inventoryImg.Fill(assets.UIBackgroundColor)

	inventory := archetype.NewTagged(w, "Inventory").
		WithParent(ui).
		WithLayer(domain.SpriteUILayerUI).
		With(component.Active).
		WithSprite(component.SpriteData{
			Image: inventoryImg,
		}).
		With(component.Collider).
		With(component.Inventory).
		Entry()

	component.Collider.SetValue(inventory, component.ColliderData{
		Width:  float64(inventoryWidth),
		Height: float64(g.screenHeight),
		Layer:  domain.CollisionLayerButton,
	})

	inventoryText := archetype.NewTagged(w, "Inventory Text").
		WithParent(inventory).
		WithLayerInherit().
		WithPosition(math.Vec2{
			X: 10,
			Y: 10,
		}).
		WithText(component.TextData{
			Text: "Inventory (e)",
		}).
		Entry()

	domain.InventoryUpdatedEvent.Subscribe(w, func(w donburi.World, event domain.InventoryUpdated) {
		text := "Inventory (e)\n\n- " + formatAsDollars(event.Money) + "\n"
		for _, item := range event.Items {
			var count string
			if item.Count > 1 {
				count = fmt.Sprintf(" x%v", item.Count)
			}
			text += fmt.Sprintf("- %v%v\n", item.Name, count)
		}
		component.Text.Get(inventoryText).Text = text
		archetype.AdjustTextWidth(inventoryText, inventoryWidth-20)
	})
}

func formatAsDollars(amount int) string {
	cents := amount % 100
	dollars := amount / 100
	return fmt.Sprintf("$%v.%02v", dollars, cents)
}

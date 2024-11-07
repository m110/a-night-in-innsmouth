package scene

import (
	"fmt"

	"github.com/m110/secrets/domain"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	donburievents "github.com/yohamta/donburi/features/events"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
	"github.com/m110/secrets/events"
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
	debug := system.NewDebug(g.loadLevel)

	g.systems = []System{
		system.NewDialog(),
		system.NewControls(),
		system.NewInventory(),
		system.NewVelocity(),
		system.NewCollision(),
		system.NewText(),
		system.NewTimeToLive(),
		system.NewDestroy(),
		debug,
	}

	g.drawables = []Drawable{
		system.NewRender(),
		debug,
	}

	g.world = g.createWorld()

	g.init()
}

func (g *Game) createWorld() donburi.World {
	world := donburi.NewWorld()

	archetype.NewCamera(world, math.Vec2{
		X: 0,
		Y: 0,
	}, engine.FloatRange{
		Min: 1,
		Max: 1,
	})

	story := domain.NewStory(world, assets.Story)

	game := world.Entry(world.Create(component.Game))
	component.Game.SetValue(game, component.GameData{
		Story: story,
		Settings: component.Settings{
			ScreenWidth:  g.screenWidth,
			ScreenHeight: g.screenHeight,
		},
	})

	world.Create(component.Debug)

	ui := archetype.NewUIRoot(world)

	archetype.NewDialog(world)
	archetype.NewPassage(world, story.PassageByTitle("Start"))

	g.createInventory(world, ui)

	story.AddMoney(1000)

	archetype.New(world).
		WithScale(math.Vec2{
			X: 0.5,
			Y: 0.5,
		}).
		WithLayer(component.SpriteLayerBackground).
		WithSprite(component.SpriteData{
			Image: assets.Background,
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

	inventoryButton := archetype.New(w).
		WithParent(ui).
		WithLayer(component.SpriteUILayerUI).
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
		Layer:  component.CollisionLayerButton,
	})

	archetype.New(w).
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

	inventory := archetype.New(w).
		WithParent(ui).
		WithLayer(component.SpriteUILayerUI).
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
		Layer:  component.CollisionLayerButton,
	})

	inventoryText := archetype.New(w).
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

	events.InventoryUpdatedEvent.Subscribe(w, func(w donburi.World, event events.InventoryUpdated) {
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

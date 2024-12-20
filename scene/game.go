package scene

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
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

type Initializer interface {
	Init(w donburi.World)
}

type Stopper interface {
	Stop(w donburi.World)
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

	switchToTitleFunc func()
}

func NewGame(screenWidth int, screenHeight int, switchToTitle func()) *Game {
	g := &Game{
		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		switchToTitleFunc: switchToTitle,
	}

	g.loadLevel()

	return g
}

func (g *Game) loadLevel() {
	debug := system.NewDebug(g.loadLevel)

	g.systems = []System{
		debug,
		system.NewControls(),
		system.NewInventory(),
		system.NewVelocity(),
		system.NewCameraFollow(),
		system.NewCollision(),
		system.NewAnimation(),
		system.NewHierarchyValidator(),
		system.NewDetectPOI(),
		system.NewText(),
		system.NewAudio(),
		system.NewTimeToLive(),
		system.NewDestroy(),
		system.NewDimensions(),
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

	story := domain.NewStory(world, assets.Assets.Story)
	settings := assets.Assets.Settings

	game := world.Entry(world.Create(component.Game, component.Input))
	component.Game.SetValue(game, component.GameData{
		Story:         story,
		Dimensions:    system.CalculateDimensions(g.screenWidth, g.screenHeight),
		SwitchToTitle: g.switchToTitle,
	})
	component.Input.SetValue(game, component.InputData{
		Disabled:      false,
		MoveRightKeys: []ebiten.Key{ebiten.KeyD, ebiten.KeyRight},
		MoveLeftKeys:  []ebiten.Key{ebiten.KeyA, ebiten.KeyLeft},
		ActionKeys:    []ebiten.Key{ebiten.KeySpace},
		MoveSpeed:     settings.Character.MoveSpeed,
	})

	ui := archetype.NewTagged(world, "UI").
		WithLayer(domain.SpriteUILayerUI).
		Entry()

	archetype.NewDialog(world)

	builder := newDebugUIBuilder(world)
	builder.Create()
	g.createInventory(world, ui)

	story.AddMoney(settings.Character.StartMoney)
	for _, i := range settings.Character.StartItems {
		story.AddItem(i)
	}

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

	archetype.ChangeLevel(world, story.StartLevel)

	// TODO a hack to not show the initial money/item received event
	domain.MoneyReceivedEvent.ProcessEvents(world)
	domain.ItemReceivedEvent.ProcessEvents(world)

	domain.MoneyReceivedEvent.Subscribe(world, func(w donburi.World, event domain.MoneyReceived) {
		paragraph := domain.Paragraph{
			Text: fmt.Sprintf("[Received %v]", formatAsDollars(event.Amount)),
			Type: domain.ParagraphTypeReceived,
		}
		archetype.AddLogParagraph(w, paragraph, archetype.ParagraphOptions{})
	})

	domain.MoneySpentEvent.Subscribe(world, func(w donburi.World, event domain.MoneySpent) {
		paragraph := domain.Paragraph{
			Text: fmt.Sprintf("[Spent %v]", formatAsDollars(event.Amount)),
			Type: domain.ParagraphTypeLost,
		}
		archetype.AddLogParagraph(w, paragraph, archetype.ParagraphOptions{})
	})

	domain.ItemReceivedEvent.Subscribe(world, func(w donburi.World, event domain.ItemReceived) {
		paragraph := domain.Paragraph{
			Text: fmt.Sprintf("[Received %v]", event.Item.Name),
			Type: domain.ParagraphTypeReceived,
		}
		archetype.AddLogParagraph(w, paragraph, archetype.ParagraphOptions{})
	})

	domain.ItemLostEvent.Subscribe(world, func(w donburi.World, event domain.ItemLost) {
		paragraph := domain.Paragraph{
			Text: fmt.Sprintf("[Lost %v]", event.Item.Name),
			Type: domain.ParagraphTypeLost,
		}
		archetype.AddLogParagraph(w, paragraph, archetype.ParagraphOptions{})
	})

	return world
}

func (g *Game) OnLayoutChange(width, height int) {
	g.screenWidth = width
	g.screenHeight = height

	if g.world == nil {
		return
	}

	gameEntry, ok := donburi.NewQuery(filter.Contains(component.Game)).First(g.world)
	if ok {
		game := component.Game.Get(gameEntry)
		game.Dimensions.ScreenWidth = width
		game.Dimensions.ScreenHeight = height
		game.Dimensions.Updated = true
	}
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

func (g *Game) createInventory(w donburi.World, ui *donburi.Entry) {
	game := component.MustFindGame(w)

	inventoryWidth := game.Dimensions.InventoryWidth
	inventoryButtonHeight := int(assets.NormalFont.Size * 2)
	marginLeft := int(float64(inventoryWidth) * 0.05)

	inventoryButtonImg := ebiten.NewImage(inventoryWidth, inventoryButtonHeight)
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
		Rect:  engine.NewRect(0, 0, float64(inventoryWidth), float64(inventoryButtonHeight)),
		Layer: domain.CollisionLayerButton,
	})

	textPos := math.Vec2{
		X: float64(marginLeft),
		Y: float64(inventoryButtonHeight-int(assets.NormalFont.Size)) / 2.0,
	}

	archetype.NewTagged(w, "Inventory Button Text").
		WithParent(inventoryButton).
		WithLayerInherit().
		WithText(component.TextData{
			Text: "Inventory (e)",
		}).
		WithPosition(textPos)

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
		Rect:  engine.NewRect(0, 0, float64(inventoryWidth), float64(g.screenHeight)),
		Layer: domain.CollisionLayerButton,
	})

	inventoryText := archetype.NewTagged(w, "Inventory Text").
		WithParent(inventory).
		WithLayerInherit().
		WithPosition(textPos).
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
		archetype.AdjustTextWidth(inventoryText, inventoryWidth-marginLeft*2)
	})
}

func formatAsDollars(amount int) string {
	cents := amount % 100
	dollars := amount / 100
	return fmt.Sprintf("$%v.%02v", dollars, cents)
}

func (g *Game) switchToTitle() {
	for _, s := range g.systems {
		if stop, ok := s.(Stopper); ok {
			stop.Stop(g.world)
		}
	}

	for _, d := range g.drawables {
		if stop, ok := d.(Stopper); ok {
			stop.Stop(g.world)
		}
	}

	g.switchToTitleFunc()
}

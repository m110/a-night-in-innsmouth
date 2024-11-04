package scene

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
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
	debug := system.NewDebug(g.loadLevel)

	g.systems = []System{
		system.NewDialog(),
		system.NewControls(),
		system.NewVelocity(),
		system.NewCollision(),
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

	story := component.NewStory(assets.Story)

	game := world.Entry(world.Create(component.Game))
	component.Game.SetValue(game, component.GameData{
		Story: story,
		Settings: component.Settings{
			ScreenWidth:  g.screenWidth,
			ScreenHeight: g.screenHeight,
		},
	})

	world.Create(component.Debug)

	archetype.NewUIRoot(world)

	archetype.NewDialog(world, story.PassageByTitle("Arkham"))

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

	events.ProcessAllEvents(g.world)
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, s := range g.drawables {
		s.Draw(g.world, screen)
	}
}

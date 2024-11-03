package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/scene"
)

type Scene interface {
	Update()
	Draw(screen *ebiten.Image)
}

type Game struct {
	scene        Scene
	screenWidth  int
	screenHeight int
}

type Config struct {
	Quick        bool
	ScreenWidth  int
	ScreenHeight int
}

func NewGame(config Config) *Game {
	assets.MustLoadAssets()

	g := &Game{
		screenWidth:  config.ScreenWidth,
		screenHeight: config.ScreenHeight,
	}

	if config.Quick {
		g.switchToGame()
	} else {
		g.switchToTitle()
	}

	return g
}

func (g *Game) switchToTitle() {
	g.scene = scene.NewTitle(g.screenWidth, g.screenHeight, g.switchToGame)
}

func (g *Game) switchToGame() {
	g.scene = scene.NewGame(g.screenWidth, g.screenHeight)
}

func (g *Game) Update() error {
	g.scene.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	if g.screenWidth == 0 || g.screenHeight == 0 {
		return width, height
	}
	return g.screenWidth, g.screenHeight
}

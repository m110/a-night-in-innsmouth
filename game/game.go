package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

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

	loadingLines []string
}

type Config struct {
	Quick        bool
	ScreenWidth  int
	ScreenHeight int
}

func NewGame(config Config) *Game {
	g := &Game{
		screenWidth:  config.ScreenWidth,
		screenHeight: config.ScreenHeight,
	}

	assets.MustLoadFonts()

	progressChan := make(chan string)
	go func() {
		assets.MustLoadAssets(progressChan)
		close(progressChan)
	}()

	go func() {
		for line := range progressChan {
			g.loadingLines = append(g.loadingLines, line)
		}

		if config.Quick {
			g.switchToGame()
		} else {
			g.switchToTitle()
		}
	}()

	return g
}

func (g *Game) switchToTitle() {
	g.scene = scene.NewTitle(g.screenWidth, g.screenHeight, g.switchToGame)
}

func (g *Game) switchToGame() {
	g.scene = scene.NewGame(g.screenWidth, g.screenHeight)
}

func (g *Game) Update() error {
	if g.scene == nil {
		return nil
	}
	g.scene.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.scene == nil {
		for i, line := range g.loadingLines {
			op := &text.DrawOptions{}
			op.LineSpacing = 1.5
			op.GeoM.Translate(10, float64(20+20*i))
			text.Draw(screen, fmt.Sprintf("%v...", line), assets.NormalFont, op)
		}
		return
	}
	g.scene.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	if g.screenWidth == 0 || g.screenHeight == 0 {
		return width, height
	}
	return g.screenWidth, g.screenHeight
}

package game

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/scene"
)

type Scene interface {
	Update()
	Draw(screen *ebiten.Image)
	OnLayoutChange(width, height int)
}

type Game struct {
	scene Scene

	rawScreenWidth  int
	rawScreenHeight int
	screenWidth     int
	screenHeight    int

	loadingLines []string
}

type Config struct {
	Quick bool
}

func NewGame(config Config) *Game {
	g := &Game{}

	assets.MustLoadFonts()
	audio.NewContext(44100)

	progressChan := make(chan string)
	errorChan := make(chan error)

	go assets.LoadAssets(progressChan, errorChan)

	go func() {
		defer func() {
			close(progressChan)
			close(errorChan)
		}()

		err := <-errorChan
		if err != nil {
			g.loadingLines = append(g.loadingLines, fmt.Sprintf("ERROR: %v", err))
			return
		}

		if config.Quick {
			g.switchToGame()
		} else {
			g.switchToTitle()
		}
	}()

	go func() {
		for line := range progressChan {
			g.loadingLines = append(g.loadingLines, fmt.Sprintf("%v...", line))
		}
	}()

	return g
}

func (g *Game) switchToTitle() {
	g.scene = scene.NewTitle(g.screenWidth, g.screenHeight, g.switchToGame)
}

func (g *Game) switchToGame() {
	g.scene = scene.NewGame(g.screenWidth, g.screenHeight, g.switchToGame)
}

func (g *Game) Update() error {
	defer func() {
		if r := recover(); r != nil {
			g.loadingLines = append(g.loadingLines, fmt.Sprintf("PANIC (Update): %v", r))
			g.scene = nil
		}
	}()

	if g.rawScreenWidth == 0 || g.rawScreenHeight == 0 {
		return nil
	}
	if g.scene == nil {
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	g.scene.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	defer func() {
		if r := recover(); r != nil {
			g.loadingLines = append(g.loadingLines, fmt.Sprintf("PANIC (Draw): %v", r))
			g.scene = nil
		}
	}()

	if g.rawScreenWidth == 0 || g.rawScreenHeight == 0 {
		return
	}

	if g.scene == nil {
		op := &text.DrawOptions{}
		op.LineSpacing = assets.NormalFont.Size
		op.GeoM.Translate(10, 10)
		lines := strings.Join(g.loadingLines, "\n")
		text.Draw(screen, lines, assets.NormalFont, op)
		return
	}
	g.scene.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	if g.rawScreenWidth != width || g.rawScreenHeight != height {
		g.rawScreenWidth = width
		g.rawScreenHeight = height

		scale := ebiten.Monitor().DeviceScaleFactor()

		g.screenWidth = int(float64(width) * scale)
		g.screenHeight = int(float64(height) * scale)

		g.loadingLines = append(g.loadingLines, fmt.Sprintf("layout change: %dx%d -> %dx%d (scale: %v)", width, height, g.screenWidth, g.screenHeight, scale))

		if g.scene != nil {
			g.scene.OnLayoutChange(g.screenWidth, g.screenHeight)
		}
	}

	return g.screenWidth, g.screenHeight
}

package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"path"
	"strings"

	"github.com/m110/secrets/engine"

	"github.com/lafriks/go-tiled"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/m110/secrets/assets/twine"
	"github.com/m110/secrets/domain"
)

var (
	//go:embed fonts/UndeadPixelLight.ttf
	normalFontData []byte

	//go:embed *
	assetsFS embed.FS

	//go:embed story.twee
	story []byte

	Story domain.RawStory

	SmallFont  *text.GoTextFace
	NormalFont *text.GoTextFace

	Character []*ebiten.Image

	Levels = map[string]domain.Level{}
)

func MustLoadAssets() {
	SmallFont = mustLoadFont(normalFontData, 10)
	NormalFont = mustLoadFont(normalFontData, 24)

	s, err := twine.ParseStory(string(story))
	if err != nil {
		panic(err)
	}
	Story = s

	characterFrames := 4
	Character = make([]*ebiten.Image, 4)
	for i := range characterFrames {
		Character[i] = mustNewEbitenImage(mustReadFile(fmt.Sprintf("character/character-%v.png", i+1)))
	}

	levelPaths, err := fs.Glob(assetsFS, "levels/*.tmx")
	if err != nil {
		panic(err)
	}

	for _, p := range levelPaths {
		name := strings.TrimSuffix(path.Base(p), ".tmx")
		Levels[name] = mustLoadLevel(p)
	}
}

func mustLoadLevel(path string) domain.Level {
	levelMap, err := tiled.LoadFile(path, tiled.WithFileSystem(assetsFS))
	if err != nil {
		panic(err)
	}

	var imageName string
	for _, t := range levelMap.ImageLayers {
		if t.Name == "Background" {
			imageName = t.Image.Source
		}
	}

	if imageName == "" {
		panic("background image not found")
	}

	var pois []domain.POI
	for _, o := range levelMap.ObjectGroups {
		for _, obj := range o.Objects {
			if obj.Class == "poi" {
				passage := obj.Properties.GetString("passage")
				assertPassageExists(passage)

				pois = append(pois, domain.POI{
					Rect:    engine.NewRect(obj.X, obj.Y, obj.Width, obj.Height),
					Passage: passage,
				})
			}
		}
	}

	return domain.Level{
		Background: mustNewEbitenImage(mustReadFile(fmt.Sprintf("levels/%v", imageName))),
		POIs:       pois,
	}
}

func assertPassageExists(name string) {
	for _, p := range Story.Passages {
		if p.Title == name {
			return
		}
	}

	panic(fmt.Sprintf("passage not found: %v", name))
}

func mustLoadFont(data []byte, size int) *text.GoTextFace {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	return &text.GoTextFace{
		Source:    s,
		Direction: text.DirectionLeftToRight,
		Size:      float64(size),
		Language:  language.English,
	}
}

func mustReadFile(name string) []byte {
	data, err := assetsFS.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return data
}

func mustNewEbitenImage(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

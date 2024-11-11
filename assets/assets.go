package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"

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

	Background *ebiten.Image

	Character []*ebiten.Image

	LevelBusStation   *ebiten.Image
	LevelHotel        *ebiten.Image
	LevelInnsmouth    *ebiten.Image
	LevelTrainStation *ebiten.Image
)

func MustLoadAssets() {
	SmallFont = mustLoadFont(normalFontData, 10)
	NormalFont = mustLoadFont(normalFontData, 24)

	s, err := twine.ParseStory(string(story))
	if err != nil {
		panic(err)
	}
	Story = s

	Background = mustNewEbitenImage(mustReadFile("background.png"))
	LevelBusStation = mustNewEbitenImage(mustReadFile("levels/bus-station.jpeg"))
	LevelHotel = mustNewEbitenImage(mustReadFile("levels/hotel.jpeg"))
	LevelInnsmouth = mustNewEbitenImage(mustReadFile("levels/innsmouth.jpeg"))
	LevelTrainStation = mustNewEbitenImage(mustReadFile("levels/train-station.jpeg"))

	characterFrames := 4
	Character = make([]*ebiten.Image, 4)
	for i := range characterFrames {
		Character[i] = mustNewEbitenImage(mustReadFile(fmt.Sprintf("character/character-%v.png", i+1)))
	}
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

package assets

import (
	"bytes"
	"embed"
	"errors"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/m110/secrets/assets/loader"
	"github.com/m110/secrets/domain"
)

var (
	//go:embed fonts/UndeadPixelLight.ttf
	normalFontData []byte

	//go:embed game/*
	assetsFS embed.FS

	SmallFont  *text.GoTextFace
	NormalFont *text.GoTextFace
	LargeFont  *text.GoTextFace

	Assets *domain.Assets
)

func MustLoadFonts() {
	UpdateFonts(14)
}

func UpdateFonts(size int) {
	SmallFont = mustLoadFont(normalFontData, int(float64(size)*0.6))
	NormalFont = mustLoadFont(normalFontData, size)
	LargeFont = mustLoadFont(normalFontData, int(float64(size)*1.4))
}

func LoadAssets(progressChan chan<- string, errorChan chan<- error) {
	if Assets != nil {
		errorChan <- errors.New("assets already loaded")
		return
	}

	assets, err := loader.LoadAssets(assetsFS, progressChan)
	if err != nil {
		errorChan <- err
		return
	}

	Assets = assets
	errorChan <- nil
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

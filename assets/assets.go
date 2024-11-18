package assets

import (
	"bytes"
	"embed"

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

	Assets *domain.Assets
)

func MustLoadAssets() {
	if Assets != nil {
		panic("assets already loaded")
	}

	assets, err := loader.LoadAssets(assetsFS)
	if err != nil {
		panic(err)
	}

	Assets = assets

	SmallFont = mustLoadFont(normalFontData, 10)
	NormalFont = mustLoadFont(normalFontData, 24)
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

package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	//go:embed fonts/UndeadPixelLight.ttf
	normalFontData []byte
	//go:embed fonts/kenney-future-narrow.ttf
	narrowFontData []byte

	//go:embed *
	assetsFS embed.FS

	SmallFont  font.Face
	NormalFont font.Face
	NarrowFont font.Face
)

func MustLoadAssets() {
	SmallFont = mustLoadFont(normalFontData, 10)
	NormalFont = mustLoadFont(normalFontData, 24)
	NarrowFont = mustLoadFont(narrowFontData, 24)
}

func mustLoadFont(data []byte, size int) font.Face {
	f, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}

	return face
}

func mustNewEbitenImage(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

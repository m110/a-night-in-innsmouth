package system

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

const (
	inventoryWidthPercent = 0.2
	dialogWidthPercent    = 0.5
	logHeightPercent      = 0.6

	// 2 margins + 2 * 4 buttons + 3 spaces between buttons
	dialogOptionRows = 13
)

type Dimensions struct{}

func NewDimensions() *Dimensions {
	return &Dimensions{}
}

func (d *Dimensions) Update(w donburi.World) {
	game := component.MustFindGame(w)

	if !game.Dimensions.Updated {
		return
	}

	dim := CalculateDimensions(game.Dimensions.ScreenWidth, game.Dimensions.ScreenHeight)
	game.Dimensions = dim
}

func CalculateDimensions(screenWidth int, screenHeight int) component.Dimensions {
	dim := component.Dimensions{
		Updated: false,

		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,

		InventoryWidth: engine.IntPercent(screenWidth, inventoryWidthPercent),

		DialogWidth:     engine.IntPercent(screenWidth, dialogWidthPercent),
		DialogLogHeight: engine.IntPercent(screenHeight, logHeightPercent),
	}

	dim.DialogOptionsHeight = screenHeight - dim.DialogLogHeight
	dim.DialogOptionsRowHeight = dim.DialogOptionsHeight / dialogOptionRows

	fontSize := int(float64(dim.DialogOptionsRowHeight) * 0.9)
	assets.UpdateFonts(fontSize)

	return dim
}

package component

import "github.com/yohamta/donburi"

type LayerID int

// SpriteLayerInherit is a special value that indicates that the entity should
// inherit the layer of its parent entity + 1.
const SpriteLayerInherit LayerID = 0

const (
	SpriteLayerBackground LayerID = 100 + iota*10
	SpriteLayerForeground
	SpriteLayerIndicator
)

const (
	SpriteUILayerBackground = 200 + iota*10
	SpriteUILayerUI
	SpriteUILayerButtons
	SpriteUILayerTop
)

type LayerData struct {
	Layer LayerID
}

var Layer = donburi.NewComponentType[LayerData]()

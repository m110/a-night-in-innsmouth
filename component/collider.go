package component

import (
	"github.com/yohamta/donburi"
)

const (
	CollisionLayerNone      ColliderLayer = iota
	CollisionLayerButton    ColliderLayer = iota
	CollisionLayerCharacter ColliderLayer = iota
	CollisionLayerPOI       ColliderLayer = iota
)

type ColliderLayer int

type ColliderData struct {
	Width  float64
	Height float64
	Layer  ColliderLayer
}

var Collider = donburi.NewComponentType[ColliderData]()

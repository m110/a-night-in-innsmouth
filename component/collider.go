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

	CollidesWith           map[CollisionKey]Collision
	JustCollidedWith       map[CollisionKey]struct{}
	JustOutOfCollisionWith map[CollisionKey]struct{}
}

type CollisionKey struct {
	Layer ColliderLayer
	Other donburi.Entity
}

type Collision struct {
	TimesSeen int
	Detected  bool
}

var Collider = donburi.NewComponentType[ColliderData]()

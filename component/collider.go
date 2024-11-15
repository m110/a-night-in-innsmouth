package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"
)

type ColliderData struct {
	Width  float64
	Height float64
	Layer  domain.ColliderLayer

	CollidesWith           map[CollisionKey]Collision
	JustCollidedWith       map[CollisionKey]struct{}
	JustOutOfCollisionWith map[CollisionKey]struct{}
}

type CollisionKey struct {
	Layer domain.ColliderLayer
	Other donburi.Entity
}

type Collision struct {
	TimesSeen int
	Detected  bool
}

var Collider = donburi.NewComponentType[ColliderData]()

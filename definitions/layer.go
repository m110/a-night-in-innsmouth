package definitions

// TODO package introduced just to avoid circular dependencies
// Figure out how to remove import domain -> events

type LayerID int

const (
	SpriteLayerBackground LayerID = 100 + iota*10
	SpriteLayerPOI
	SpriteLayerCharacter
	SpriteLayerForeground
	SpriteLayerIndicator
)

const (
	SpriteUILayerUI = 200 + iota*10
	SpriteUILayerBackground
	SpriteUILayerText
	SpriteUILayerButtons
	SpriteUILayerTop
)

type ColliderLayer int

const (
	CollisionLayerNone      ColliderLayer = iota
	CollisionLayerButton    ColliderLayer = iota
	CollisionLayerCharacter ColliderLayer = iota
	CollisionLayerPOI       ColliderLayer = iota
)

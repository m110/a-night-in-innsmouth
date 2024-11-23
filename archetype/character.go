package archetype

import (
	"time"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

func NewCharacter(parent *donburi.Entry, scale float64, movementBounds component.MovementBoundsData) *donburi.Entry {
	// Sanity check
	if scale == 0 {
		panic("character scale cannot be 0")
	}

	w := parent.World
	character := NewTagged(w, "Character").
		WithScale(math.Vec2{
			X: scale,
			Y: scale,
		}).
		WithParent(parent).
		WithLayer(domain.SpriteLayerCharacter).
		WithSprite(component.SpriteData{
			Image: assets.Assets.Character.Frames[2],
		}).
		With(component.Velocity).
		WithSpriteBounds().
		With(component.Collider).
		With(component.Animator).
		With(component.Character).
		With(component.MovementBounds).
		Entry()

	colliderRect := assets.Assets.Character.Collider.Scale(scale)

	component.Collider.SetValue(character, component.ColliderData{
		Rect:  colliderRect,
		Layer: domain.CollisionLayerCharacter,
	})

	sprite := component.Sprite.Get(character)
	frames := assets.Assets.Character.Frames

	currentFrame := 0

	anim := component.Animator.Get(character)

	anim.SetAnimation("walk", &component.Animation{
		Timer: engine.NewTimer(150 * time.Millisecond),
		Update: func(e *donburi.Entry, a *component.Animation) {
			if a.Timer.IsReady() {
				currentFrame++
				if currentFrame >= len(frames) {
					currentFrame = 0
				}
				sprite.Image = frames[currentFrame]
				a.Timer.Reset()
			}
		},
		OnStart: func(e *donburi.Entry) {
			currentFrame = 0
			sprite.Image = frames[0]
		},
		OnStop: func(e *donburi.Entry) {
			currentFrame = 0
			sprite.Image = frames[2]
		},
	})

	r := movementBounds.Range
	r.Max -= colliderRect.Width
	movementBounds.Range = r

	component.MovementBounds.SetValue(character, movementBounds)

	return character
}

func RotateCharacterTowards(target *donburi.Entry) {
	character, ok := engine.FindWithComponent(target.World, component.Character)
	if !ok {
		return
	}

	pos := transform.WorldPosition(character)
	targetPos := transform.WorldPosition(target)

	sprite := component.Sprite.Get(character)

	if pos.X < targetPos.X {
		sprite.FlipY = false
	} else {
		sprite.FlipY = true
	}
}

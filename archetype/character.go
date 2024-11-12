package archetype

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewCharacter(parent *donburi.Entry) *donburi.Entry {
	w := parent.World
	character := NewTagged(w, "Character").
		WithPosition(math.Vec2{
			X: 450,
			Y: 1000,
		}).
		WithScale(math.Vec2{
			X: 0.4,
			Y: 0.4,
		}).
		WithParent(parent).
		WithLayer(component.SpriteLayerCharacter).
		WithSprite(component.SpriteData{
			Image: assets.Character[2],
		}).
		With(component.Input).
		With(component.Velocity).
		WithSpriteBounds().
		WithBoundsAsCollider(component.CollisionLayerCharacter).
		With(component.Animation).
		With(component.Character).
		Entry()

	component.Input.SetValue(character, component.InputData{
		Disabled:     false,
		MoveRightKey: ebiten.KeyD,
		MoveLeftKey:  ebiten.KeyA,
		ActionKey:    ebiten.KeySpace,
		MoveSpeed:    6,
	})

	sprite := component.Sprite.Get(character)
	frames := []*ebiten.Image{
		assets.Character[0],
		assets.Character[1],
		assets.Character[2],
		assets.Character[3],
	}

	currentFrame := 0

	anim := component.Animation.Get(character)

	component.Animation.SetValue(character, component.AnimationData{
		Timer: engine.NewTimer(200 * time.Millisecond),
		Update: func(e *donburi.Entry) {
			if anim.Timer.IsReady() {
				currentFrame++
				if currentFrame >= len(frames) {
					currentFrame = 0
				}
				sprite.Image = frames[currentFrame]
				anim.Timer.Reset()
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

	return character
}

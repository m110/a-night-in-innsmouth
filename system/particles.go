package system

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/yohamta/donburi/features/transform"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type Particles struct {
	query *donburi.Query
}

func NewParticles() *Particles {
	return &Particles{
		query: donburi.NewQuery(
			filter.Contains(
				component.Particle,
			),
		),
	}
}

func (s *Particles) Update(w donburi.World) {
	s.query.Each(w, func(entry *donburi.Entry) {
		particle := component.Particle.Get(entry)
		particle.Life -= 1
		if particle.Life <= 0 {
			component.Destroy(entry)
		}

		particle.Color.A = uint8(255 * (particle.Life / particle.MaxLife))
	})

	if rand.Float64() < 0.3 {
		entry := w.Entry(w.Create(transform.Transform, component.Velocity, component.Particle))

		// Random position near center
		t := transform.GetTransform(entry)
		t.LocalPosition.X = 320 + (rand.Float64()*40 - 20)
		t.LocalPosition.Y = 240 + (rand.Float64()*40 - 20)

		// Random velocity
		vel := component.Velocity.Get(entry)
		angle := rand.Float64() * math.Pi * 2
		speed := 0.1 + rand.Float64()*0.1
		vel.Velocity.X = math.Cos(angle) * speed
		vel.Velocity.Y = math.Sin(angle) * speed

		// Initialize particle properties
		particle := component.Particle.Get(entry)
		particle.Life = 180
		particle.MaxLife = 180
		particle.Size = 2 + rand.Float64()*2
		particle.Color = color.RGBA{
			R: 255,
			G: uint8(rand.Float64() * 100), // Slight variation in color
			B: uint8(rand.Float64() * 50),
			A: 255,
		}
	}
}

func (s *Particles) Draw(w donburi.World, screen *ebiten.Image) {
	s.query.Each(w, func(entry *donburi.Entry) {
		pos := transform.WorldPosition(entry)
		particle := component.Particle.Get(entry)

		// Draw particle
		ebitenutil.DrawCircle(
			screen,
			pos.X,
			pos.Y,
			particle.Size,
			particle.Color,
		)
	})
}

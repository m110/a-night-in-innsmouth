package system

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"
	"golang.org/x/image/colornames"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/component"
)

type Debug struct {
	query     *donburi.Query
	debug     *component.DebugData
	offscreen *ebiten.Image

	pausedCameraVelocity math.Vec2

	restartLevelCallback func()
}

func NewDebug(restartLevelCallback func()) *Debug {
	return &Debug{
		query: donburi.NewQuery(
			filter.Contains(transform.Transform, component.Sprite),
		),
		// TODO figure out the proper size
		offscreen:            ebiten.NewImage(3000, 3000),
		restartLevelCallback: restartLevelCallback,
	}
}

func (d *Debug) Update(w donburi.World) {
	if d.debug == nil {
		debug, ok := donburi.NewQuery(filter.Contains(component.Debug)).First(w)
		if !ok {
			return
		}

		d.debug = component.Debug.Get(debug)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySlash) {
		d.debug.Enabled = !d.debug.Enabled
	}
}

func (d *Debug) Draw(w donburi.World, screen *ebiten.Image) {
	if d.debug == nil || !d.debug.Enabled {
		return
	}

	allCount := w.Len()

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Entities: %v", allCount), 0, 0)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %v", ebiten.ActualTPS()), 0, 20)

	d.offscreen.Clear()
	d.query.Each(w, func(entry *donburi.Entry) {
		t := transform.Transform.Get(entry)
		sprite := component.Sprite.Get(entry)

		position := transform.WorldPosition(entry)

		w, h := sprite.Image.Size()
		halfW, halfH := float64(w)/2, float64(h)/2

		x := position.X
		y := position.Y

		switch sprite.Pivot {
		case component.SpritePivotCenter:
			x -= halfW
			y -= halfH
		}

		ebitenutil.DrawRect(d.offscreen, t.LocalPosition.X-2, t.LocalPosition.Y-2, 4, 4, colornames.Lime)
		ebitenutil.DebugPrintAt(d.offscreen, fmt.Sprintf("%v", entry.Entity().Id()), int(x), int(y))
		ebitenutil.DebugPrintAt(d.offscreen, fmt.Sprintf("pos: %.0f, %.0f", position.X, position.Y), int(x), int(y)+40)
		ebitenutil.DebugPrintAt(d.offscreen, fmt.Sprintf("rot: %.0f", transform.WorldRotation(entry)), int(x), int(y)+60)

		length := 50.0
		right := position.Add(transform.Right(entry).MulScalar(length))
		up := position.Add(transform.Up(entry).MulScalar(length))

		ebitenutil.DrawLine(d.offscreen, position.X, position.Y, right.X, right.Y, colornames.Blue)
		ebitenutil.DrawLine(d.offscreen, position.X, position.Y, up.X, up.Y, colornames.Lime)

		if entry.HasComponent(component.Collider) {
			collider := component.Collider.Get(entry)
			ebitenutil.DrawLine(d.offscreen, x, y, x+collider.Width, y, colornames.Lime)
			ebitenutil.DrawLine(d.offscreen, x, y, x, y+collider.Height, colornames.Lime)
			ebitenutil.DrawLine(d.offscreen, x+collider.Width, y, x+collider.Width, y+collider.Height, colornames.Lime)
			ebitenutil.DrawLine(d.offscreen, x, y+collider.Height, x+collider.Width, y+collider.Height, colornames.Lime)
		}
	})

	camera := archetype.MustFindCamera(w)
	cameraPos := transform.Transform.Get(camera).LocalPosition
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cameraPos.X, -cameraPos.Y)
	screen.DrawImage(d.offscreen, op)
}

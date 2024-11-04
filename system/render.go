package system

import (
	"fmt"
	"image/color"
	stdmath "math"
	"sort"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

type Render struct {
	query   *donburi.Query
	uiQuery *donburi.Query

	mainBoardOffscreen *ebiten.Image
	uiOffscreen        *ebiten.Image
	game               *component.GameData

	cameraTransform *transform.TransformData

	debug *component.DebugData
}

func NewRender() *Render {
	return &Render{
		query: donburi.NewQuery(
			filter.And(
				filter.Or(
					filter.Contains(
						component.Sprite,
					),
					filter.Contains(
						component.Text,
					),
				),
				filter.Not(
					filter.Contains(component.UI),
				),
			),
		),
		uiQuery: donburi.NewQuery(
			filter.And(
				filter.Or(
					filter.Contains(
						component.Sprite,
					),
					filter.Contains(
						component.Text,
					),
				),
				filter.Contains(component.UI),
			),
		),
	}
}

func (r *Render) Init(w donburi.World) {
	r.game = component.MustFindGame(w)

	camera := archetype.MustFindCamera(w)
	r.cameraTransform = transform.GetTransform(camera)

	cam := component.Camera.Get(camera)

	imageWidth := int(float64(r.game.Settings.ScreenWidth) / cam.Zoom.Min)
	imageHeight := int(float64(r.game.Settings.ScreenHeight) / cam.Zoom.Min)
	r.mainBoardOffscreen = ebiten.NewImage(imageWidth, imageHeight)

	r.uiOffscreen = ebiten.NewImage(r.game.Settings.ScreenWidth, r.game.Settings.ScreenHeight)

	r.debug = component.Debug.Get(engine.MustFindWithComponent(w, component.Debug))
}

func (r *Render) Draw(w donburi.World, screen *ebiten.Image) {
	cameraPos := r.cameraTransform.LocalPosition
	cameraScale := r.cameraTransform.LocalScale

	r.mainBoardOffscreen.Clear()
	r.uiOffscreen.Clear()

	var count, uiCount int
	byLayer := map[int][]entryWithLayer{}

	r.query.Each(w, func(entry *donburi.Entry) {
		layer := component.Layer.Get(entry).Layer
		byLayer[int(layer)] = append(byLayer[int(layer)], entryWithLayer{
			entry: entry,
			layer: layer,
			ui:    false,
		})
		count++
	})

	r.uiQuery.Each(w, func(entry *donburi.Entry) {
		layer := component.Layer.Get(entry).Layer
		byLayer[int(layer)] = append(byLayer[int(layer)], entryWithLayer{
			entry: entry,
			layer: layer,
			ui:    true,
		})
		uiCount++
	})

	renderSprite := func(entry *donburi.Entry, ui bool) {
		if !entry.HasComponent(component.Sprite) {
			return
		}

		sprite := component.Sprite.Get(entry)

		if sprite.Image == nil {
			panic(fmt.Sprintf("sprite image is nil: %s", entry))
		}

		if sprite.Hidden {
			return
		}

		position := transform.WorldPosition(entry)

		offscreen := r.mainBoardOffscreen
		if ui {
			offscreen = r.uiOffscreen
		} else {
			position.X -= cameraPos.X
			position.Y -= cameraPos.Y
		}

		bounds := sprite.Image.Bounds()

		width, height := bounds.Dx(), bounds.Dy()

		halfW, halfH := float64(width)/2, float64(height)/2

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-halfW, -halfH)
		op.GeoM.Rotate(float64(int(transform.WorldRotation(entry)-sprite.OriginalRotation)%360) * 2 * stdmath.Pi / 360)
		op.GeoM.Translate(halfW, halfH)

		x := position.X
		y := position.Y

		scale := transform.WorldScale(entry)

		switch sprite.Pivot {
		case component.SpritePivotCenter:
			x -= halfW * scale.X
			y -= halfH * scale.Y
		}

		op.GeoM.Scale(scale.X, scale.Y)

		if sprite.AlphaOverride != nil {
			op.ColorM.Scale(1.0, 1.0, 1.0, sprite.AlphaOverride.A)
		}
		if sprite.ColorOverride != nil {
			op.ColorM.Translate(sprite.ColorOverride.R, sprite.ColorOverride.G, sprite.ColorOverride.B, 0)
		}

		op.GeoM.Translate(x, y)

		offscreen.DrawImage(sprite.Image, op)
	}

	renderText := func(entry *donburi.Entry, ui bool) {
		if !entry.HasComponent(component.Text) {
			return
		}

		t := component.Text.Get(entry)

		if t.Hidden {
			return
		}

		font := archetype.FontFromSize(t.Size)

		pos := transform.WorldPosition(entry)

		var col color.Color = assets.TextColor
		if t.Color != nil {
			col = t.Color
		}

		length := utf8.RuneCountInString(t.Text)
		if t.Streaming {
			length = int(float64(length) * t.StreamingTimer.PercentDone())
		}

		textToDraw := t.Text[:length]

		op := &text.DrawOptions{}
		op.LineSpacing = 24
		op.GeoM.Translate(pos.X, pos.Y)
		op.ColorScale.ScaleWithColor(col)

		offscreen := r.mainBoardOffscreen
		if ui {
			offscreen = r.uiOffscreen
		}

		text.Draw(offscreen, textToDraw, font, op)
	}

	var layers []int
	for l := range byLayer {
		layers = append(layers, l)
	}

	sort.Ints(layers)

	for _, layer := range layers {
		for _, entry := range byLayer[layer] {
			if !isActive(entry.entry) {
				return
			}

			if entry.ui {
				renderSprite(entry.entry, true)
				renderText(entry.entry, true)
			} else {
				renderSprite(entry.entry, false)
				renderText(entry.entry, false)
			}
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cameraScale.X, cameraScale.Y)
	screen.DrawImage(r.mainBoardOffscreen, op)

	op = &ebiten.DrawImageOptions{}
	screen.DrawImage(r.uiOffscreen, op)

	if r.debug.Enabled {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %v", int(ebiten.ActualFPS())), 10, 10)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %v", int(ebiten.ActualTPS())), 10, 30)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered: %v", count), 10, 70)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered UI: %v", uiCount), 10, 90)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("World entities: %v", w.Len()), 10, 130)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera: (%v, %v)", cameraPos.X, cameraPos.Y), 10, 160)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Camera scale: (%v, %v)", cameraScale.X, cameraScale.Y), 10, 180)
	}
}

func isActive(entry *donburi.Entry) bool {
	if entry.HasComponent(component.Active) && !component.Active.Get(entry).Active {
		return false
	}

	for {
		parent, ok := transform.GetParent(entry)
		if !ok {
			break
		}
		if parent.HasComponent(component.Active) {
			if !component.Active.Get(parent).Active {
				return false
			}
		}
		entry = parent
	}

	return true
}

type entryWithLayer struct {
	entry *donburi.Entry
	layer component.LayerID
	ui    bool
}

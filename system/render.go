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
	camerasQuery *donburi.OrderedQuery[component.CameraData]

	offscreen *ebiten.Image

	debug *component.DebugData
}

func NewRender() *Render {
	return &Render{
		camerasQuery: donburi.NewOrderedQuery[component.CameraData](
			filter.Contains(
				component.Camera,
			),
		),
	}
}

func (r *Render) Init(w donburi.World) {
	game := component.MustFindGame(w)

	imageWidth := game.Settings.ScreenWidth
	imageHeight := game.Settings.ScreenHeight
	r.offscreen = ebiten.NewImage(imageWidth, imageHeight)

	r.debug = component.Debug.Get(engine.MustFindWithComponent(w, component.Debug))
}

func (r *Render) Draw(w donburi.World, screen *ebiten.Image) {
	r.offscreen.Clear()

	count := 0
	r.camerasQuery.EachOrdered(w, component.Camera, func(entry *donburi.Entry) {
		camera := component.Camera.Get(entry)

		if !camera.Root.HasComponent(component.Layer) {
			panic("missing root layer")
		}

		rootLayer := component.Layer.Get(camera.Root).Layer
		children := getAllChildren(camera.Root, rootLayer)

		byLayer := map[int][]entryWithLayer{}
		for _, child := range children {
			byLayer[int(child.layer)] = append(byLayer[int(child.layer)], child)
		}

		var layers []int
		for l := range byLayer {
			layers = append(layers, l)
		}

		sort.Ints(layers)

		camera.Viewport.Clear()

		for _, layer := range layers {
			for _, child := range byLayer[layer] {
				count++

				if child.entry.HasComponent(component.Sprite) {
					renderSprite(child.entry, camera.Viewport)
				}

				if child.entry.HasComponent(component.Text) {
					renderText(child.entry, camera.Viewport)
				}
			}
		}

		if camera.Mask != nil {
			op := &ebiten.DrawImageOptions{}
			op.Blend = ebiten.BlendDestinationIn
			camera.Viewport.DrawImage(camera.Mask, op)
		}

		cameraPos := transform.WorldPosition(entry)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cameraPos.X, cameraPos.Y)
		r.offscreen.DrawImage(camera.Viewport, op)
	})

	screen.DrawImage(r.offscreen, nil)

	if r.debug.Enabled {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %v", int(ebiten.ActualFPS())), 10, 10)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %v", int(ebiten.ActualTPS())), 10, 30)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered entities: %v", count), 10, 150)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("World entities: %v", w.Len()), 10, 130)
	}
}

func getAllChildren(entry *donburi.Entry, rootLayer component.LayerID) []entryWithLayer {
	if !entry.Valid() || !isActive(entry) {
		return nil
	}

	result := make([]entryWithLayer, 0, 32)
	seen := make(map[*donburi.Entry]bool)

	stack := []*donburi.Entry{entry}

	for len(stack) > 0 {
		e := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !e.Valid() || !isActive(e) || seen[e] {
			continue
		}

		seen[e] = true

		if e.HasComponent(component.Sprite) || e.HasComponent(component.Text) {
			result = append(result, getEntryWithLayer(e, rootLayer))
		}

		children, ok := transform.GetChildren(e)
		if ok {
			stack = append(stack, children...)
		}
	}

	return result
}

func getEntryWithLayer(entry *donburi.Entry, rootLayer component.LayerID) entryWithLayer {
	if !entry.HasComponent(component.Layer) {
		return entryWithLayer{
			entry: entry,
			layer: rootLayer,
		}
	}

	layer := component.Layer.Get(entry)
	return entryWithLayer{
		entry: entry,
		layer: layer.Layer,
	}
}

func renderSprite(entry *donburi.Entry, offscreen *ebiten.Image) {
	sprite := component.Sprite.Get(entry)

	if sprite.Image == nil {
		panic(fmt.Sprintf("sprite image is nil: %s", entry))
	}

	if sprite.Hidden {
		return
	}

	bounds := sprite.Image.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	halfW, halfH := float64(width)/2, float64(height)/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-halfW, -halfH)
	op.GeoM.Rotate(float64(int(transform.WorldRotation(entry)-sprite.OriginalRotation)%360) * 2 * stdmath.Pi / 360)
	op.GeoM.Translate(halfW, halfH)

	position := transform.WorldPosition(entry)
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

func renderText(entry *donburi.Entry, offscreen *ebiten.Image) {
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
	op.PrimaryAlign = t.Align
	op.GeoM.Translate(pos.X, pos.Y)
	op.ColorScale.ScaleWithColor(col)

	text.Draw(offscreen, textToDraw, font, op)
}

func isActive(entry *donburi.Entry) bool {
	if entry.HasComponent(component.Active) && !component.Active.Get(entry).Active {
		return false
	}

	return true
}

type entryWithLayer struct {
	entry *donburi.Entry
	layer component.LayerID
}

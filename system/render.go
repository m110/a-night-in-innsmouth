package system

import (
	"fmt"
	"image/color"
	stdmath "math"
	"sort"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
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
				filter.Contains(
					transform.Transform,
					component.Sprite,
				),
				filter.Not(
					filter.Contains(component.UI),
				),
			),
		),
		uiQuery: donburi.NewQuery(
			filter.Contains(
				transform.Transform,
				component.UI,
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

	cameraWidth := float64(r.game.Settings.ScreenWidth) / cameraScale.X
	cameraHeight := float64(r.game.Settings.ScreenHeight) / cameraScale.Y

	cameraEdgeLeft := cameraPos.X
	cameraEdgeRight := cameraPos.X + cameraWidth
	cameraEdgeTop := cameraPos.Y
	cameraEdgeBottom := cameraPos.Y + cameraHeight

	var chunks, count, uiCount int
	worldByLayer := map[int][]*donburi.Entry{}
	worldTextByLayer := map[int][]*donburi.Entry{}
	uiByLayer := map[int][]*donburi.Entry{}
	uiTextByLayer := map[int][]*donburi.Entry{}

	r.query.Each(w, func(entry *donburi.Entry) {
		for _, child := range findChildrenWithComponent(entry, component.Sprite, component.SpriteLayerInherit) {
			l := int(child.layer)
			worldByLayer[l] = append(worldByLayer[l], child.entry)
			count++
		}

		for _, child := range findChildrenWithComponent(entry, component.Text, component.SpriteLayerInherit) {
			l := int(child.layer)
			worldTextByLayer[l] = append(worldTextByLayer[l], child.entry)
			count++
		}
	})

	r.uiQuery.Each(w, func(entry *donburi.Entry) {
		parentLayer := component.SpriteLayerInherit
		if entry.HasComponent(component.Sprite) {
			parentLayer = component.Layer.Get(entry).Layer
			l := int(parentLayer)
			uiByLayer[l] = append(uiByLayer[l], entry)
			uiCount++
		}

		for _, child := range findChildrenWithComponent(entry, component.Sprite, parentLayer) {
			l := int(child.layer)
			uiByLayer[l] = append(uiByLayer[l], child.entry)
			uiCount++
		}

		for _, child := range findChildrenWithComponent(entry, component.Text, parentLayer) {
			l := int(child.layer)
			uiTextByLayer[l] = append(uiTextByLayer[l], child.entry)
			uiCount++
		}
	})

	renderEntry := func(entry *donburi.Entry, ui bool) {
		sprite := component.Sprite.Get(entry)

		if sprite.Image == nil {
			panic(fmt.Sprintf("sprite image is nil: %s", entry))
		}

		if sprite.Hidden {
			return
		}

		if !isActive(entry) {
			return
		}

		position := transform.WorldPosition(entry)

		offscreen := r.mainBoardOffscreen
		if ui {
			offscreen = r.uiOffscreen
		} else {
			if position.X < cameraEdgeLeft ||
				position.X > cameraEdgeRight ||
				position.Y < cameraEdgeTop ||
				position.Y > cameraEdgeBottom {
				return
			}

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
		if !isActive(entry) {
			return
		}

		t := component.Text.Get(entry)

		if t.Hidden {
			return
		}

		font := assets.NormalFont
		switch t.Size {
		case component.TextSizeL:
		case component.TextSizeM:
			font = assets.NormalFont
		case component.TextSizeS:
			font = assets.SmallFont
		}

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

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pos.X, pos.Y)
		op.ColorScale.ScaleWithColor(col)

		offscreen := r.mainBoardOffscreen
		if ui {
			offscreen = r.uiOffscreen
		}

		text.DrawWithOptions(offscreen, textToDraw, font, op)
	}

	var layers []int
	for l := range worldByLayer {
		layers = append(layers, l)
	}
	for l := range worldTextByLayer {
		layers = append(layers, l)
	}
	for l := range uiByLayer {
		layers = append(layers, l)
	}
	for l := range uiTextByLayer {
		layers = append(layers, l)
	}

	sort.Ints(layers)

	for _, layer := range layers {
		for _, entry := range worldByLayer[layer] {
			renderEntry(entry, false)
		}
		for _, entry := range worldTextByLayer[layer] {
			renderText(entry, false)
		}
		for _, entry := range uiByLayer[layer] {
			renderEntry(entry, true)
		}
		for _, entry := range uiTextByLayer[layer] {
			renderText(entry, true)
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(cameraScale.X, cameraScale.Y)
	screen.DrawImage(r.mainBoardOffscreen, op)

	/*
		timeUniform := float32(time.Now().UnixNano()) / float32(time.Second) / 10
		b := r.mainBoardOffscreen.Bounds()
		options := &ebiten.DrawRectShaderOptions{}
		options.Images[0] = r.mainBoardOffscreen
		options.Uniforms = map[string]interface{}{
			"Time": timeUniform,
		}
		screen.DrawRectShader(b.Dx(), b.Dy(), assets.ShaderDistortion, options)
	*/

	op = &ebiten.DrawImageOptions{}
	screen.DrawImage(r.uiOffscreen, op)

	if r.debug.Enabled {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %v", int(ebiten.ActualFPS())), 10, 10)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TPS: %v", int(ebiten.ActualTPS())), 10, 30)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered: %v", count), 10, 70)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered UI: %v", uiCount), 10, 90)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Rendered chunks: %v", chunks), 10, 110)
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
}

func findChildrenWithComponent(e *donburi.Entry, componentType donburi.IComponentType, parentLayer component.LayerID) []entryWithLayer {
	if !e.Valid() {
		return nil
	}

	children, ok := transform.GetChildren(e)
	if !ok {
		return nil
	}

	parentLayer++

	var result []entryWithLayer
	for _, child := range children {
		if !child.Valid() {
			continue
		}

		childLayer := parentLayer

		if child.HasComponent(component.Layer) {
			overrideLayer := component.Layer.Get(child).Layer
			if overrideLayer != component.SpriteLayerInherit {
				childLayer = overrideLayer
			}
		}

		if child.HasComponent(componentType) {
			result = append(result, entryWithLayer{
				entry: child,
				layer: childLayer,
			})
		}

		result = append(result, findChildrenWithComponent(child, componentType, childLayer)...)
	}

	return result
}

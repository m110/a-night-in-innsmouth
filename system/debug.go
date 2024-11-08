package system

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"
	"golang.org/x/image/colornames"

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

	if d.debug.Enabled {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			PrintHierarchy(w)
		}
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

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(d.offscreen, op)
}

func PrintHierarchy(w donburi.World) {
	fmt.Println("\n=== Full Entity Hierarchy ===")

	// Find all entities with Transform
	roots := make([]*donburi.Entry, 0)
	donburi.NewQuery(filter.Contains(transform.Transform)).Each(w, func(entry *donburi.Entry) {
		if entry.HasComponent(transform.Transform) {
			// Only include entities without parents as roots
			if parent, ok := transform.GetParent(entry); !ok || !parent.Valid() {
				roots = append(roots, entry)
			}
		}
	})

	// Recursive print function
	var printEntry func(entry *donburi.Entry, depth int)
	printEntry = func(entry *donburi.Entry, depth int) {
		indent := strings.Repeat("  ", depth)

		if !entry.Valid() {
			fmt.Printf("%sEntity %v (invalid)\n", entry.Entity(), indent)
			return
		}
		// Print entity info
		tag := "no tag"
		if entry.HasComponent(component.Tag) {
			tag = component.Tag.Get(entry).Tag
		}

		// List relevant components
		components := []string{}
		if entry.HasComponent(component.Passage) {
			components = append(components, "Passage")
		}
		if entry.HasComponent(component.StackedView) {
			components = append(components, "StackedView")
		}
		if entry.HasComponent(component.DialogOption) {
			components = append(components, "DialogOption")
		}

		componentStr := ""
		if len(components) > 0 {
			componentStr = fmt.Sprintf(" [%s]", strings.Join(components, ", "))
		}

		fmt.Printf("%sEntity %v (%s)%s\n", indent, entry.Entity(), tag, componentStr)

		// Print children recursively
		if children, ok := transform.GetChildren(entry); ok && len(children) > 0 {
			for _, child := range children {
				printEntry(child, depth+1)
			}
		}
	}

	// Print each root
	for _, root := range roots {
		printEntry(root, 0)
	}
}

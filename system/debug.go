package system

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
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
		restartLevelCallback: restartLevelCallback,
	}
}

func (d *Debug) Init(w donburi.World) {
	game := component.MustFindGame(w)

	imageWidth := game.Settings.ScreenWidth
	imageHeight := game.Settings.ScreenHeight
	d.offscreen = ebiten.NewImage(imageWidth, imageHeight)
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

		game := engine.MustFindWithComponent(w, component.Game)
		in := component.Input.Get(game)
		if d.debug.Enabled {
			in.MoveSpeed *= 3
		} else {
			in.MoveSpeed /= 3
		}
	}

	if d.debug.Enabled {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			PrintHierarchy(w)
		}
	}
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

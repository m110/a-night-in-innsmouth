package system

import (
	"fmt"
	"strings"
	"time"

	"github.com/ebitenui/ebitenui/input"

	"github.com/m110/secrets/domain"

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
	query *donburi.Query

	pausedCameraVelocity math.Vec2
	restartLevelCallback func()

	// 0 for short press, 1 for long
	clickedSequence    []int
	longClickTimer     *engine.Timer
	betweenClicksTimer *engine.Timer
}

func NewDebug(restartLevelCallback func()) *Debug {
	return &Debug{
		query: donburi.NewQuery(
			filter.Contains(component.DebugUI),
		),

		restartLevelCallback: restartLevelCallback,
		longClickTimer:       engine.NewTimer(1 * time.Second),
		betweenClicksTimer:   engine.NewTimer(500 * time.Millisecond),
	}
}

func (d *Debug) Update(w donburi.World) {
	game := component.MustFindGame(w)

	var clicked bool
	var released bool

	d.longClickTimer.Update()
	d.betweenClicksTimer.Update()

	pressedTouchIDs := inpututil.AppendJustPressedTouchIDs(nil)
	if len(pressedTouchIDs) == 1 {
		clicked = true
	}

	releasedTouchIDs := inpututil.AppendJustReleasedTouchIDs(nil)
	if len(releasedTouchIDs) == 1 {
		released = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		clicked = true
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		released = true
	}

	var toggleDebug bool
	if released {
		if d.betweenClicksTimer.IsReady() {
			d.clickedSequence = []int{}
			d.betweenClicksTimer.Reset()
		}

		if d.longClickTimer.IsReady() {
			d.clickedSequence = append(d.clickedSequence, 1)
		} else {
			d.clickedSequence = append(d.clickedSequence, 0)
		}

		if len(d.clickedSequence) == 3 {
			if d.clickedSequence[0] == 1 && d.clickedSequence[1] == 0 && d.clickedSequence[2] == 0 {
				toggleDebug = true
			}
		}
	}

	if clicked {
		d.longClickTimer.Reset()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySlash) || toggleDebug {
		game.Debug.Enabled = !game.Debug.Enabled

		var speedChange float64
		if game.Debug.Enabled {
			speedChange = 10
		} else {
			speedChange = -10
		}
		domain.CharacterSpeedChangedEvent.Publish(w, domain.CharacterSpeedChanged{
			SpeedChange: speedChange,
		})
	}

	if game.Debug.Enabled {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			PrintHierarchy(w)
		}

		d.query.Each(w, func(entry *donburi.Entry) {
			component.DebugUI.Get(entry).UI.Update()
		})

		game.Debug.UIHovered = input.UIHovered
	}
}

func (d *Debug) Draw(w donburi.World, screen *ebiten.Image) {
	game := component.MustFindGame(w)
	if !game.Debug.Enabled {
		return
	}

	d.query.Each(w, func(entry *donburi.Entry) {
		component.DebugUI.Get(entry).UI.Draw(screen)
	})
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

package system

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/transform"
	"github.com/yohamta/donburi/filter"

	"github.com/m110/secrets/component"
)

type HierarchyValidator struct {
	lastCheck     time.Time
	checkInterval time.Duration
}

func NewHierarchyValidator() *HierarchyValidator {
	return &HierarchyValidator{
		checkInterval: time.Second, // Adjust as needed
	}
}

func (h *HierarchyValidator) Update(w donburi.World) {
	now := time.Now()
	if now.Sub(h.lastCheck) < h.checkInterval {
		return
	}
	h.lastCheck = now

	problems := DetectMultipleParents(w)
	if len(problems) > 0 {
		// Option 1: Just log
		log.Printf("WARNING: Found entities with multiple parents")
		for entity, parents := range problems {
			log.Printf("Entity %v has multiple parents:", describeEntity(entity))
			for _, parent := range parents {
				log.Printf("  - %v", describeEntity(parent))
			}
		}
	}
}

func DetectMultipleParents(world donburi.World) map[*donburi.Entry][]*donburi.Entry {
	// Map of entity -> all its parents
	multiParents := make(map[*donburi.Entry][]*donburi.Entry)

	// Helper to collect all entries in the world that might be part of hierarchies
	entries := make([]*donburi.Entry, 0)
	donburi.NewQuery(filter.Contains(transform.Transform)).Each(world, func(entry *donburi.Entry) {
		entries = append(entries, entry)
	})

	// For each entry, check its children and record parent relationships
	for _, entry := range entries {
		children, ok := transform.GetChildren(entry)
		if !ok {
			continue
		}

		for _, child := range children {
			multiParents[child] = append(multiParents[child], entry)
		}
	}

	// Filter to only entities with multiple parents
	result := make(map[*donburi.Entry][]*donburi.Entry)
	for entity, parents := range multiParents {
		if len(parents) > 1 {
			result[entity] = parents
		}
	}

	return result
}

// Helper to describe an entity (adapt this to your components)
func describeEntity(entry *donburi.Entry) string {
	parts := []string{fmt.Sprintf("%v ID: %v", entry, entry.Entity())}

	// Add component information
	if entry.HasComponent(component.Sprite) {
		parts = append(parts, "Sprite")
	}
	if entry.HasComponent(component.Text) {
		text := component.Text.Get(entry)
		parts = append(parts, fmt.Sprintf("Text: %q", text.Text))
	}
	// Add any other identifying components...

	return strings.Join(parts, ", ")
}

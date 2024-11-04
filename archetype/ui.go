package archetype

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/component"
	"github.com/m110/secrets/engine"
)

func NewUIRoot(w donburi.World) {
	New(w).
		With(component.UI)
}

func MustFindUIRoot(w donburi.World) *donburi.Entry {
	return engine.MustFindWithComponent(w, component.UI)
}

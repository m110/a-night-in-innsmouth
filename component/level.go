package component

import "github.com/yohamta/donburi"

type LevelData struct {
	Name     string
	Changing bool
}

var Level = donburi.NewComponentType[LevelData]()

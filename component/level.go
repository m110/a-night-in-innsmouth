package component

import "github.com/yohamta/donburi"

type LevelData struct {
	Name string
}

var Level = donburi.NewComponentType[LevelData]()

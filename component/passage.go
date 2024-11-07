package component

import (
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/domain"
)

type PassageData struct {
	Passage      *domain.Passage
	ActiveOption int
	Height       float64
}

var Passage = donburi.NewComponentType[PassageData]()

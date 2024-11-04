package archetype

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/assets"

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

func AdjustTextWidth(entry *donburi.Entry, width int) {
	txt := component.Text.Get(entry)
	font := FontFromSize(txt.Size)
	var newText strings.Builder
	lines := strings.Split(txt.Text, "\n")

	for i, line := range lines {
		words := strings.Fields(line)
		if len(words) == 0 {
			// Handle empty lines
			newText.WriteRune('\n')
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			// Add space only if we're adding to existing line
			newLine := currentLine + " " + word
			w, _ := text.Measure(newLine, font, 0)

			if int(w) > width {
				// Write current line and start new one
				newText.WriteString(currentLine)
				newText.WriteRune('\n')
				currentLine = word
			} else {
				currentLine = newLine
			}
		}

		// Write the last line
		newText.WriteString(currentLine)

		// Add newline if it's not the last line
		if i < len(lines)-1 {
			newText.WriteRune('\n')
		}
	}

	txt.Text = newText.String()
}

func FontFromSize(size component.TextSize) *text.GoTextFace {
	font := assets.NormalFont

	switch size {
	case component.TextSizeL:
	case component.TextSizeM:
		font = assets.NormalFont
	case component.TextSizeS:
		font = assets.SmallFont
	}

	return font
}

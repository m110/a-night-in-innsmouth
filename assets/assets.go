package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/lafriks/go-tiled"
	"github.com/yohamta/donburi/features/math"
	"golang.org/x/text/language"

	"github.com/m110/secrets/assets/twine"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

var (
	//go:embed fonts/UndeadPixelLight.ttf
	normalFontData []byte

	//go:embed *
	assetsFS embed.FS

	//go:embed story.twee
	story []byte

	Story domain.RawStory

	SmallFont  *text.GoTextFace
	NormalFont *text.GoTextFace

	Character []*ebiten.Image

	levelNames = map[string]struct{}{}
	Levels     = map[string]domain.Level{}
)

func MustLoadAssets() {
	SmallFont = mustLoadFont(normalFontData, 10)
	NormalFont = mustLoadFont(normalFontData, 24)

	s, err := twine.ParseStory(string(story))
	if err != nil {
		panic(err)
	}
	Story = s

	characterFrames := 4
	Character = make([]*ebiten.Image, 4)
	for i := range characterFrames {
		Character[i] = mustNewEbitenImage(mustReadFile(fmt.Sprintf("character/character-%v.png", i+1)))
	}

	mustLoadLevels()

	for _, l := range Levels {
		for _, p := range l.POIs {
			if p.Level != nil {
				assertLevelExists(*p.Level)
			}
		}
	}

	for _, p := range Story.Passages {
		for _, l := range p.Links {
			if l.Level != nil {
				assertLevelExists(*l.Level)
			}
		}
	}
}

func mustLoadLevels() {
	levelPaths, err := fs.Glob(assetsFS, "levels/*.tmx")
	if err != nil {
		panic(err)
	}

	for _, p := range levelPaths {
		name := strings.TrimSuffix(path.Base(p), ".tmx")
		levelNames[name] = struct{}{}
	}

	for _, p := range levelPaths {
		name := strings.TrimSuffix(path.Base(p), ".tmx")
		Levels[name] = mustLoadLevel(p)
	}
}

func mustLoadLevel(levelPath string) domain.Level {
	levelMap, err := tiled.LoadFile(levelPath, tiled.WithFileSystem(assetsFS))
	if err != nil {
		panic(err)
	}

	tilesetImages := map[uint32]*ebiten.Image{}
	for _, ts := range levelMap.Tilesets {
		if ts.Image != nil {
			// Only collection of images supported for now
			continue
		}

		for _, tile := range ts.Tiles {
			if tile.Image != nil && tile.Image.Source != "" {
				p := path.Join("levels", path.Dir(ts.Source), tile.Image.Source)
				img := mustReadFile(p)
				globalID := ts.FirstGID + tile.ID
				tilesetImages[globalID] = mustNewEbitenImage(img)
			}
		}
	}

	if len(levelMap.ImageLayers) != 1 {
		panic("expected exactly one image layer")
	}

	background := mustNewEbitenImage(mustReadFile(fmt.Sprintf("levels/%v", levelMap.ImageLayers[0].Image.Source)))

	var objects []domain.Object
	var pois []domain.POI
	var entrypoints []domain.Entrypoint
	var characterScale float64
	for _, o := range levelMap.ObjectGroups {
		for _, obj := range o.Objects {
			if obj.Class == "object" {
				img, ok := tilesetImages[obj.GID]
				if !ok {
					panic(fmt.Sprintf("object not found: %v", obj.GID))
				}

				objImg := ebiten.NewImageFromImage(img)
				bounds := objImg.Bounds()
				layer := o.Properties.GetInt("layer")

				domainObj := domain.Object{
					Image: objImg,
					Position: math.Vec2{
						X: obj.X,
						Y: obj.Y - obj.Height,
					},
					Scale: math.Vec2{
						X: obj.Width / float64(bounds.Dx()),
						Y: obj.Height / float64(bounds.Dy()),
					},
					Layer: domain.LayerID(layer),
				}

				objects = append(objects, domainObj)
			}

			if obj.Class == "poi" {
				y := obj.Y
				var objectImg *ebiten.Image
				if obj.GID != 0 {
					// Image-based objects have pivot set to bottom-left
					// Other objects have pivot set to top-left
					y -= obj.Height
					img, ok := tilesetImages[obj.GID]
					if !ok {
						panic(fmt.Sprintf("object not found: %v", obj.GID))
					}
					objectImg = ebiten.NewImageFromImage(img)
				}

				var domainEdge *domain.Direction
				edge := domain.Direction(obj.Properties.GetString("edge"))
				if edge != "" {
					if edge != domain.EdgeLeft && edge != domain.EdgeRight {
						panic(fmt.Sprintf("invalid edge: %v", edge))
					}

					domainEdge = &edge
				}

				rect := engine.NewRect(obj.X, y, obj.Width, obj.Height)
				poi := domain.POI{
					ID:           fmt.Sprint(obj.ID),
					Image:        objectImg,
					TriggerRect:  rect,
					Rect:         rect,
					EdgeTrigger:  domainEdge,
					TouchTrigger: obj.Properties.GetBool("touchTrigger"),
				}

				passage := obj.Properties.GetString("passage")
				if passage != "" {
					assertPassageExists(passage)
					poi.Passage = passage
				}

				level := obj.Properties.GetString("level")
				if level != "" {
					parts := strings.Split(level, ",")
					var entrypoint *int
					if len(parts) == 2 {
						e, err := strconv.Atoi(strings.TrimSpace(parts[1]))
						if err != nil {
							panic(err)
						}
						entrypoint = &e
					} else if len(parts) > 2 {
						panic(fmt.Sprintf("invalid level: %v", level))
					}

					poi.Level = &domain.TargetLevel{
						Name:       strings.TrimSpace(parts[0]),
						Entrypoint: entrypoint,
					}
				}

				if passage == "" && level == "" {
					panic(fmt.Sprintf("poi has no passage or level: %v", obj.ID))
				}

				if passage != "" && level != "" {
					panic(fmt.Sprintf("poi has both passage and level: %v", obj.ID))
				}

				pois = append(pois, poi)
			}

			if obj.Class == "entrypoint" {
				pos := domain.CharacterPosition{
					LocalPosition: math.Vec2{
						X: obj.X,
						Y: obj.Y,
					},
				}

				if obj.GID != 0 {
					// Image-based entrypoints have pivot set to bottom-left
					// Other entrypoints have pivot set to top-left
					pos.LocalPosition.Y -= obj.Height
				}

				if obj.Properties.GetBool("flipY") {
					pos.FlipY = true
				}

				entrypoint := domain.Entrypoint{
					Index:             obj.Properties.GetInt("index"),
					CharacterPosition: pos,
				}

				// The first entrypoint's scale is used for all entrypoint
				if entrypoint.Index == 0 {
					characterScale = obj.Height / float64(Character[2].Bounds().Dy())
				}

				entrypoints = append(entrypoints, entrypoint)
			}
		}
	}

	sort.Slice(entrypoints, func(i, j int) bool {
		return entrypoints[i].Index < entrypoints[j].Index
	})

	for i, e := range entrypoints {
		if e.Index != i {
			panic(fmt.Sprintf("entrypoint index is not sequential: %v", e.Index))
		}
	}

	for _, o := range levelMap.ObjectGroups {
		for _, obj := range o.Objects {
			if obj.Class == "trigger" {
				rect := engine.NewRect(obj.X, obj.Y, obj.Width, obj.Height)
				poiID := obj.Properties.GetString("poi")

				var found bool
				for i, p := range pois {
					if poiID == p.ID {
						p.TriggerRect = rect
						pois[i] = p
						found = true
						break
					}
				}

				if !found {
					panic(fmt.Sprintf("poi not found: %v", poiID))
				}
			}
		}
	}

	var startPassage string
	var cameraZoom float64
	if levelMap.Properties != nil {
		startPassage = levelMap.Properties.GetString("startPassage")
		assertPassageExists(startPassage)

		cameraZoom = levelMap.Properties.GetFloat("cameraZoom")
	}

	return domain.Level{
		Background:     background,
		POIs:           pois,
		Objects:        objects,
		StartPassage:   startPassage,
		Entrypoints:    entrypoints,
		CameraZoom:     cameraZoom,
		CharacterScale: characterScale,
	}
}

func assertPassageExists(name string) {
	for _, p := range Story.Passages {
		if p.Title == name {
			return
		}
	}

	panic(fmt.Sprintf("passage not found: %v", name))
}

func assertLevelExists(level domain.TargetLevel) {
	if _, ok := Levels[level.Name]; !ok {
		panic(fmt.Sprintf("level not found: %v"))
	}

	if level.Entrypoint != nil {
		if *level.Entrypoint < 0 || *level.Entrypoint >= len(Levels[level.Name].Entrypoints) {
			panic(fmt.Sprintf("entrypoint not found: %v %v", level.Name, *level.Entrypoint))
		}
	}
}

func mustLoadFont(data []byte, size int) *text.GoTextFace {
	s, err := text.NewGoTextFaceSource(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	return &text.GoTextFace{
		Source:    s,
		Direction: text.DirectionLeftToRight,
		Size:      float64(size),
		Language:  language.English,
	}
}

func mustReadFile(name string) []byte {
	data, err := assetsFS.ReadFile(name)
	if err != nil {
		panic(err)
	}

	return data
}

func mustNewEbitenImage(data []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

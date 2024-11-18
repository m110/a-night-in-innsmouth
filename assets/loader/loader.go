package loader

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/assets/twine"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

var (
	levelNames = map[string]struct{}{}
)

func LoadAssets(assetsFS fs.FS) (*domain.Assets, error) {
	twee, err := fs.ReadFile(assetsFS, "game/game.twee")
	if err != nil {
		return nil, err
	}

	story, err := twine.ParseStory(string(twee))
	if err != nil {
		return nil, err
	}

	characterFrames := 4
	character := make([]*ebiten.Image, 4)
	for i := range characterFrames {
		charFile, err := fs.ReadFile(assetsFS, fmt.Sprintf("game/character/character-%v.png", i+1))
		if err != nil {
			return nil, err
		}

		character[i], err = newImageFromBytes(charFile)
		if err != nil {
			return nil, err
		}
	}

	characterHeight := float64(character[2].Bounds().Dy())
	levels, err := loadLevels(assetsFS, characterHeight)
	if err != nil {
		return nil, err
	}

	for _, l := range levels {
		if l.StartPassage != "" {
			err = assertPassageExists(story, l.StartPassage)
			if err != nil {
				return nil, err
			}
		}

		for _, p := range l.POIs {
			if p.Passage != "" {
				err = assertPassageExists(story, p.Passage)
				if err != nil {
					return nil, err
				}
			}
			if p.Level != nil {
				err = assertLevelExists(levels, *p.Level)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	for _, p := range story.Passages {
		for _, l := range p.Links {
			if l.Level != nil {
				err = assertLevelExists(levels, *l.Level)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &domain.Assets{
		Story:     story,
		Levels:    levels,
		Character: character,
	}, nil
}

// TODO Passing characterHeight here seems weird, reconsider?
func loadLevels(assetsFS fs.FS, characterHeight float64) (map[string]domain.Level, error) {
	levelPaths, err := fs.Glob(assetsFS, "game/levels/*.tmx")
	if err != nil {
		return nil, err
	}

	for _, p := range levelPaths {
		name := strings.TrimSuffix(path.Base(p), ".tmx")
		levelNames[name] = struct{}{}
	}

	levels := map[string]domain.Level{}
	for _, p := range levelPaths {
		name := strings.TrimSuffix(path.Base(p), ".tmx")
		l, err := loadLevel(assetsFS, p, characterHeight)
		if err != nil {
			return nil, err
		}
		levels[name] = l
	}

	return levels, nil
}

func loadLevel(assetsFS fs.FS, levelPath string, characterHeight float64) (domain.Level, error) {
	levelMap, err := tiled.LoadFile(levelPath, tiled.WithFileSystem(assetsFS))
	if err != nil {
		return domain.Level{}, err
	}

	tilesetImages := map[uint32]*ebiten.Image{}
	for _, ts := range levelMap.Tilesets {
		if ts.Image != nil {
			// Only collection of images supported for now
			continue
		}

		for _, tile := range ts.Tiles {
			if tile.Image != nil && tile.Image.Source != "" {
				p := path.Join("game/levels", path.Dir(ts.Source), tile.Image.Source)
				imgBytes, err := fs.ReadFile(assetsFS, p)
				if err != nil {
					return domain.Level{}, err
				}

				globalID := ts.FirstGID + tile.ID
				img, err := newImageFromBytes(imgBytes)
				if err != nil {
					return domain.Level{}, err
				}
				tilesetImages[globalID] = img
			}
		}
	}

	if len(levelMap.ImageLayers) != 1 {
		return domain.Level{}, fmt.Errorf("expected one image layer, got: %v", len(levelMap.ImageLayers))
	}

	bgBytes, err := fs.ReadFile(assetsFS, path.Join("game/levels", levelMap.ImageLayers[0].Image.Source))
	if err != nil {
		return domain.Level{}, err
	}

	background, err := newImageFromBytes(bgBytes)
	if err != nil {
		return domain.Level{}, err
	}

	var objects []domain.Object
	var pois []domain.POI
	var entrypoints []domain.Entrypoint
	var characterScale float64
	var limits *engine.FloatRange

	for _, o := range levelMap.ObjectGroups {
		for _, obj := range o.Objects {
			if obj.Class == "limits" {
				if limits != nil {
					return domain.Level{}, errors.New("multiple limits objects")
				}

				limits = &engine.FloatRange{
					Min: obj.X,
					Max: obj.X + obj.Width,
				}
			}

			if obj.Class == "object" {
				img, ok := tilesetImages[obj.GID]
				if !ok {
					return domain.Level{}, errors.New(fmt.Sprintf("object not found: %v", obj.GID))
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
						return domain.Level{}, fmt.Errorf("object not found: %v", obj.GID)
					}
					objectImg = ebiten.NewImageFromImage(img)
				}

				var domainEdge *domain.Direction
				edge := domain.Direction(obj.Properties.GetString("edge"))
				if edge != "" {
					if edge != domain.EdgeLeft && edge != domain.EdgeRight {
						return domain.Level{}, fmt.Errorf("invalid edge: %v", edge)
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
					poi.Passage = passage
				}

				level := obj.Properties.GetString("level")
				if level != "" {
					parts := strings.Split(level, ",")
					var entrypoint *int
					if len(parts) == 2 {
						e, err := strconv.Atoi(strings.TrimSpace(parts[1]))
						if err != nil {
							return domain.Level{}, err
						}
						entrypoint = &e
					} else if len(parts) > 2 {
						return domain.Level{}, fmt.Errorf("invalid level: %v", level)
					}

					poi.Level = &domain.TargetLevel{
						Name:       strings.TrimSpace(parts[0]),
						Entrypoint: entrypoint,
					}
				}

				if passage == "" && level == "" {
					return domain.Level{}, fmt.Errorf("poi has no passage or level: %v", obj.ID)
				}

				if passage != "" && level != "" {
					return domain.Level{}, fmt.Errorf("poi has both passage and level: %v", obj.ID)
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
					characterScale = obj.Height / characterHeight
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
			return domain.Level{}, fmt.Errorf("entrypoint index is not sequential: %v", e.Index)
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
					return domain.Level{}, fmt.Errorf("poi not found: %v", poiID)
				}
			}
		}
	}

	var startPassage string
	var cameraZoom float64
	if levelMap.Properties != nil {
		startPassage = levelMap.Properties.GetString("startPassage")
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
		Limits:         limits,
	}, nil
}

func assertPassageExists(story domain.RawStory, name string) error {
	for _, p := range story.Passages {
		if p.Title == name {
			return nil
		}
	}

	return fmt.Errorf("passage not found: %v", name)
}

func assertLevelExists(levels map[string]domain.Level, level domain.TargetLevel) error {
	if _, ok := levels[level.Name]; !ok {
		return fmt.Errorf("level not found: %v", level.Name)
	}

	if level.Entrypoint != nil {
		if *level.Entrypoint < 0 || *level.Entrypoint >= len(levels[level.Name].Entrypoints) {
			return fmt.Errorf("entrypoint not found: %v %v", level.Name, *level.Entrypoint)
		}
	}

	return nil
}

func newImageFromBytes(data []byte) (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return ebiten.NewImageFromImage(img), nil
}

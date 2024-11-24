package loader

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/lafriks/go-tiled"
	"github.com/yohamta/donburi/features/math"

	"github.com/m110/secrets/assets/twine"
	"github.com/m110/secrets/domain"
	"github.com/m110/secrets/engine"
)

var (
	levelNames = map[string]struct{}{}

	// TODO could refactor into struct instead of a global variable
	characterCollider *engine.Rect
)

func LoadAssets(assetsFS fs.FS, progressChan chan<- string) (*domain.Assets, error) {
	progressChan <- "Parsing story"

	twee, err := fs.ReadFile(assetsFS, "game/game.twee")
	if err != nil {
		return nil, err
	}

	story, err := twine.ParseStory(string(twee))
	if err != nil {
		return nil, err
	}

	progressChan <- "Loading character"

	characterFrames := 5
	character := make([]*ebiten.Image, characterFrames)
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

	progressChan <- "Loading levels"
	characterHeight := float64(character[2].Bounds().Dy())
	levels, err := loadLevels(assetsFS, characterHeight, progressChan)
	if err != nil {
		return nil, err
	}

	progressChan <- "Loading audio"
	sounds, err := loadSounds(assetsFS)
	if err != nil {
		return nil, err
	}

	music, err := loadMusic(assetsFS)
	if err != nil {
		return nil, err
	}

	progressChan <- "Validating assets"

	assets := &domain.Assets{
		Story:  story,
		Levels: levels,
		Character: domain.Character{
			Frames:   character,
			Collider: *characterCollider,
		},
		Sounds: sounds,
		Music:  music,
	}

	err = validateAssets(assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func validateAssets(assets *domain.Assets) error {
	for _, l := range assets.Levels {
		// Levels without character should have a passage, otherwise the game will be stuck
		if len(l.Entrypoints) == 0 {
			err := assertPassageExists(assets.Story, l.Name)
			if err != nil {
				return err
			}
		}

		for _, p := range l.POIs {
			if p.Passage != "" {
				err := assertPassageExists(assets.Story, p.Passage)
				if err != nil {
					return err
				}
			}
			if p.Level != nil {
				err := assertLevelExists(assets.Levels, *p.Level)
				if err != nil {
					return err
				}
			}
		}
	}

	availableItems := map[string]struct{}{}
	availableFacts := map[string]struct{}{}

	for _, p := range assets.Story.Passages {
		for _, m := range p.Macros {
			if m.Type == domain.MacroTypePlayMusic && m.Value != "" {
				if _, ok := assets.Music[m.Value]; !ok {
					return fmt.Errorf("music not found: %v", m.Value)
				}
			}

			if m.Type == domain.MacroTypeAddItem {
				if m.Value == "" {
					return errors.New("empty item name")
				}
				availableItems[m.Value] = struct{}{}
			}

			if m.Type == domain.MacroTypeSetFact {
				if m.Value == "" {
					return errors.New("empty fact name")
				}
				availableFacts[m.Value] = struct{}{}
			}
		}

		for _, l := range p.Links {
			if l.Level != nil {
				err := assertLevelExists(assets.Levels, *l.Level)
				if err != nil {
					return err
				}
			}
		}
	}

	requiredItems := map[string]struct{}{}
	requiredFacts := map[string]struct{}{}

	for _, p := range assets.Story.Passages {
		for _, c := range p.Conditions {
			if c.Type == domain.ConditionTypeHasItem {
				requiredItems[c.Value] = struct{}{}
			}

			if c.Type == domain.ConditionTypeFact {
				requiredFacts[c.Value] = struct{}{}
			}
		}

		for _, m := range p.Links {
			for _, c := range m.Conditions {
				if c.Type == domain.ConditionTypeHasItem {
					requiredItems[c.Value] = struct{}{}
				}

				if c.Type == domain.ConditionTypeFact {
					requiredFacts[c.Value] = struct{}{}
				}
			}
		}
	}

	for item := range requiredItems {
		if _, ok := availableItems[item]; !ok {
			return fmt.Errorf("required item not found: %v", item)
		}
	}

	for fact := range requiredFacts {
		if _, ok := availableFacts[fact]; !ok {
			return fmt.Errorf("required fact not found: %v", fact)
		}
	}

	if characterCollider == nil {
		return errors.New("character collider not found")
	}

	return nil
}

// TODO Passing characterHeight here seems weird, reconsider?
func loadLevels(assetsFS fs.FS, characterHeight float64, progressChan chan<- string) (map[string]domain.Level, error) {
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
		progressChan <- fmt.Sprintf("Loading level %v", name)
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

	levelName := strings.TrimSuffix(path.Base(levelPath), ".tmx")

	tilesetImages := map[uint32]*ebiten.Image{}
	for _, ts := range levelMap.Tilesets {
		if ts.Image != nil {
			// Only collection of images supported for now
			continue
		}

		for _, tile := range ts.Tiles {
			if tile.Image != nil && tile.Image.Source != "" {
				// TODO Probably not the best approach here
				if characterCollider == nil && ts.Name == "Character" {
					for _, og := range tile.ObjectGroups {
						for _, o := range og.Objects {
							if o.Class == "collider" {
								r := engine.NewRect(o.X, o.Y, o.Width, o.Height)
								characterCollider = &r
							}
						}
					}
				}

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

	bgLayer := levelMap.ImageLayers[0]
	if bgLayer.OffsetX != 0 || bgLayer.OffsetY != 0 {
		return domain.Level{}, errors.New("background layer offset is not (0,0)")
	}

	bgPath := path.Join("game/levels", bgLayer.Image.Source)

	_, err = fs.Stat(assetsFS, bgPath)
	if err != nil {
		return domain.Level{}, err
	}

	loadBackground := func() *ebiten.Image {
		bgBytes, err := fs.ReadFile(assetsFS, bgPath)
		if err != nil {
			panic(err)
		}
		background, err := newImageFromBytes(bgBytes)
		if err != nil {
			panic(err)
		}

		return background
	}

	var objects []domain.Object
	var pois []domain.POI
	var entrypoints []domain.Entrypoint
	var character domain.LevelCharacter
	var limits *engine.FloatRange
	var fadepoint *math.Vec2

	for _, o := range levelMap.ObjectGroups {
		for _, obj := range o.Objects {
			if obj.Class == "fadepoint" {
				if fadepoint != nil {
					return domain.Level{}, errors.New("multiple fadepoint objects")
				}

				fadepoint = &math.Vec2{
					X: obj.X,
					Y: obj.Y,
				}
			}
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
					return domain.Level{}, errors.New(fmt.Sprintf("object not found: %v on level %v", obj.GID, levelPath))
				}

				objImg := ebiten.NewImageFromImage(img)
				bounds := objImg.Bounds()
				// TODO Replace by separate layers
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
				object := domain.Object{
					Position: math.Vec2{
						X: obj.X,
						Y: obj.Y,
					},
				}

				if obj.GID != 0 {
					// Image-based objects have pivot set to bottom-left
					// Other objects have pivot set to top-left
					object.Position.Y -= obj.Height
					img, ok := tilesetImages[obj.GID]
					if !ok {
						return domain.Level{}, fmt.Errorf("object not found: %v", obj.GID)
					}
					object.Image = ebiten.NewImageFromImage(img)

					object.Scale = math.Vec2{
						X: obj.Width / float64(img.Bounds().Dx()),
						Y: obj.Height / float64(img.Bounds().Dy()),
					}
				}

				var domainEdge *domain.Direction
				edge := domain.Direction(obj.Properties.GetString("edge"))
				if edge != "" {
					if edge != domain.EdgeLeft && edge != domain.EdgeRight {
						return domain.Level{}, fmt.Errorf("invalid edge: %v", edge)
					}

					domainEdge = &edge
				}

				rect := engine.NewRect(object.Position.X, object.Position.Y, obj.Width, obj.Height)
				poi := domain.POI{
					ID:           fmt.Sprint(obj.ID),
					Object:       object,
					TriggerRect:  rect,
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
					},
				}

				if obj.Properties.GetBool("flipY") {
					pos.FlipY = true
				}

				entrypoint := domain.Entrypoint{
					Index:             obj.Properties.GetInt("index"),
					CharacterPosition: pos,
				}

				// The first entrypoint's scale and Y position is used for all entrypoints
				if entrypoint.Index == 0 {
					character.Scale = obj.Height / characterHeight
					character.PosY = obj.Y

					if obj.GID != 0 {
						// Image-based entrypoints have pivot set to bottom-left
						// Other entrypoints have pivot set to top-left
						character.PosY -= obj.Height
					}
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

	var cameraZoom float64
	if levelMap.Properties != nil {
		cameraZoom = levelMap.Properties.GetFloat("cameraZoom")
	}

	return domain.Level{
		Name:        levelName,
		Background:  loadBackground,
		POIs:        pois,
		Objects:     objects,
		Entrypoints: entrypoints,
		CameraZoom:  cameraZoom,
		Character:   character,
		Limits:      limits,
		Fadepoint:   fadepoint,
	}, nil
}

func loadSounds(assetsFS fs.FS) (domain.Sounds, error) {
	sounds := domain.Sounds{}

	var err error
	sounds.Click1, err = loadMP3Stream(assetsFS, "audio/sounds/click-1.mp3")
	if err != nil {
		return sounds, err
	}

	return sounds, nil
}

func loadMusic(assetsFS fs.FS) (map[string][]byte, error) {
	music := map[string][]byte{}

	files, err := fs.ReadDir(assetsFS, "audio/music")
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fileName := f.Name()
		ext := path.Ext(fileName)
		if ext == ".mp3" {
			data, err := loadMP3Stream(assetsFS, path.Join("audio/music", fileName))
			if err != nil {
				return nil, err
			}

			name := strings.TrimSuffix(fileName, ext)
			music[name] = data
		} else {
			return nil, fmt.Errorf("unsupported music format: %v", fileName)
		}
	}

	return music, nil
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

func loadMP3Stream(fileSystem fs.FS, path string) ([]byte, error) {
	data, err := fs.ReadFile(fileSystem, path)
	if err != nil {
		return nil, err
	}

	stream, err := mp3.DecodeWithoutResampling(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	streamBytes, err := io.ReadAll(stream)
	if err != nil {
		return nil, err
	}

	return streamBytes, nil
}

package scene

import (
	image2 "image"
	"image/color"
	"sort"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
)

type debugUIBuilder struct {
	world donburi.World
	ui    *ebitenui.UI

	buttonImage   *widget.ButtonImage
	checkboxImage *widget.CheckboxGraphicImage

	factCheckboxes map[string]*widget.Checkbox
}

func newDebugUIBuilder(world donburi.World) *debugUIBuilder {
	buttonImg := image.NewNineSliceColor(color.RGBA{0, 0, 0, 128})
	hoverImg := image.NewNineSliceColor(color.RGBA{0, 50, 0, 128})

	checkboxSize := 40

	uncheckedImage := newSquareImage(checkboxSize, color.White)
	checkedImage := newSquareImage(checkboxSize, color.NRGBA{0, 255, 0, 255})
	greyedImage := newSquareImage(checkboxSize, color.NRGBA{255, 0, 0, 255})

	return &debugUIBuilder{
		ui:    &ebitenui.UI{},
		world: world,
		buttonImage: &widget.ButtonImage{
			Idle:         buttonImg,
			Hover:        hoverImg,
			Pressed:      hoverImg,
			PressedHover: hoverImg,
			Disabled:     buttonImg,
		},
		checkboxImage: &widget.CheckboxGraphicImage{
			Unchecked: &widget.ButtonImageImage{
				Idle: uncheckedImage,
			},
			Checked: &widget.ButtonImageImage{
				Idle: checkedImage,
			},
			Greyed: &widget.ButtonImageImage{
				Idle: greyedImage,
			},
		},
		factCheckboxes: map[string]*widget.Checkbox{},
	}
}

func (b *debugUIBuilder) Create() {
	debugUI := b.world.Entry(b.world.Create(component.DebugUI))
	component.DebugUI.SetValue(debugUI, component.DebugUIData{
		UI: b.ui,
	})

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	levelButtonsContainer := b.newLevelButtonsContainer()
	rootContainer.AddChild(levelButtonsContainer)

	dockContainer := b.newDockContainer()
	rootContainer.AddChild(dockContainer)

	b.ui.Container = rootContainer
}

func (b *debugUIBuilder) newLevelButtonsContainer() *widget.Container {
	var levels []string
	for _, level := range assets.Assets.Levels {
		levels = append(levels, level.Name)
	}
	sort.Strings(levels)

	container := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(2, 2),
			widget.GridLayoutOpts.Stretch([]bool{true}, []bool{}),
		)),
		// Position the button container in the top-right corner
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
			Padding:            widget.Insets{Left: 50},
		})),
	)

	// Create buttons for each level
	for _, levelName := range levels {
		button := widget.NewButton(
			widget.ButtonOpts.Image(b.buttonImage),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				}),
			),
			widget.ButtonOpts.Text(levelName, assets.NormalFont, &widget.ButtonTextColor{
				Idle: color.White,
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   5,
				Right:  5,
				Top:    2,
				Bottom: 2,
			}),
			widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
				targetLevel := domain.TargetLevel{
					Name: levelName,
				}

				// If any entry point exists, pick the first one
				if len(assets.Assets.Levels[levelName].Entrypoints) > 0 {
					ep := 0
					targetLevel.Entrypoint = &ep
				}

				archetype.ChangeLevel(b.world, targetLevel)
			}),
		)
		container.AddChild(button)
	}

	return container
}

func (b *debugUIBuilder) newDockContainer() *widget.Container {
	characterWindow := b.newCharacterWindow()
	factsWindow := b.newFactsWindow()

	dockContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(2),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch([]bool{true, true}, []bool{}),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
			Padding: widget.Insets{
				Left:   50,
				Bottom: 50,
			},
		})),
	)

	buttons := []struct {
		text    string
		handler func(*widget.ButtonClickedEventArgs)
	}{
		{
			text: "Character",
			handler: func(args *widget.ButtonClickedEventArgs) {
				x, y := characterWindow.Contents.PreferredSize()
				r := image2.Rect(0, 0, x, y)
				r = r.Add(image2.Point{200, 200})
				characterWindow.SetLocation(r)
				b.ui.AddWindow(characterWindow)
			},
		},
		{
			text: "Facts",
			handler: func(args *widget.ButtonClickedEventArgs) {
				x, y := factsWindow.Contents.PreferredSize()
				r := image2.Rect(0, 0, x, y)
				r = r.Add(image2.Point{200, 200})
				factsWindow.SetLocation(r)
				b.ui.AddWindow(factsWindow)
			},
		},
	}

	for _, btn := range buttons {
		button := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(),
			widget.ButtonOpts.Image(b.buttonImage),
			widget.ButtonOpts.Text(btn.text, assets.NormalFont, &widget.ButtonTextColor{
				Idle: color.White,
			}),
			widget.ButtonOpts.TextPadding(widget.Insets{
				Left:   10,
				Right:  10,
				Top:    5,
				Bottom: 5,
			}),
			widget.ButtonOpts.ClickedHandler(btn.handler),
		)
		dockContainer.AddChild(button)
	}

	return dockContainer
}

func (b *debugUIBuilder) newCharacterWindow() *widget.Window {
	windowContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.RGBA{0, 0, 0, 255})),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Spacing(2),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    4,
				Bottom: 4,
				Left:   4,
				Right:  4,
			}),
		)),
	)

	titleContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{150, 150, 150, 255})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	titleContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("Character", assets.SmallFont, color.NRGBA{254, 255, 255, 255}),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	))

	speedSlider := widget.NewSlider(
		widget.SliderOpts.MinMax(0, 100),
		widget.SliderOpts.HandleImage(b.buttonImage),
	)
	windowContainer.AddChild(speedSlider)

	debugWindow := widget.NewWindow(
		widget.WindowOpts.TitleBar(titleContainer, 24),
		widget.WindowOpts.Draggable(),
		widget.WindowOpts.MinSize(200, 100),
		widget.WindowOpts.Contents(windowContainer),
	)

	return debugWindow
}

func (b *debugUIBuilder) newFactsWindow() *widget.Window {
	game := component.MustFindGame(b.world)
	var facts []string
	for fact := range game.Story.Facts {
		facts = append(facts, fact)
	}

	sort.Strings(facts)

	container := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.RGBA{0, 0, 0, 255})),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch: true,
			}),
		),
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(2),
			widget.RowLayoutOpts.Padding(widget.Insets{
				Top:    5,
				Bottom: 25,
				Left:   4,
				Right:  4,
			}),
		)),
	)

	titleContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(image.NewNineSliceColor(color.NRGBA{150, 150, 150, 255})),
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)
	titleContainer.AddChild(widget.NewText(
		widget.TextOpts.Text("Facts", assets.SmallFont, color.NRGBA{254, 255, 255, 255}),
		widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionCenter,
			VerticalPosition:   widget.AnchorLayoutPositionCenter,
		})),
	))

	for _, fact := range facts {
		row := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Spacing(4),
				widget.RowLayoutOpts.Padding(widget.Insets{
					Top:    5,
					Bottom: 5,
				}),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Stretch: true,
				}),
			),
		)

		factText := widget.NewText(
			widget.TextOpts.Text(fact, assets.NormalFont, color.White),
			widget.TextOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Stretch:  true,
				Position: widget.RowLayoutPositionCenter,
			})),
		)

		state := widget.WidgetUnchecked
		if game.Story.Facts[fact] {
			state = widget.WidgetChecked
		}

		checkbox := widget.NewCheckbox(
			widget.CheckboxOpts.ButtonOpts(
				widget.ButtonOpts.Image(b.buttonImage),
				widget.ButtonOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.RowLayoutData{
						Position: widget.RowLayoutPositionCenter,
					}),
				),
			),
			widget.CheckboxOpts.Image(b.checkboxImage),
			widget.CheckboxOpts.InitialState(state),
			widget.CheckboxOpts.StateChangedHandler(func(args *widget.CheckboxChangedEventArgs) {
				if args.State == widget.WidgetChecked {
					game.Story.Facts[fact] = true
				} else {
					game.Story.Facts[fact] = false
				}
			}),
		)

		b.factCheckboxes[fact] = checkbox

		row.AddChild(checkbox)
		row.AddChild(factText)

		container.AddChild(row)
	}

	domain.StoryFactSetEvent.Subscribe(b.world, func(w donburi.World, event domain.StoryFactSet) {
		checkbox, ok := b.factCheckboxes[event.Fact]
		if !ok {
			// This can happen if the fact is not used in any condition in the story
			// No point displaying the checkbox then, as it has no effect
			return
		}
		checkbox.SetState(widget.WidgetChecked)
	})

	window := widget.NewWindow(
		widget.WindowOpts.TitleBar(titleContainer, 24),
		widget.WindowOpts.Draggable(),
		widget.WindowOpts.Contents(container),
	)

	return window
}

func newSquareImage(size int, c color.Color) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	img.Fill(c)
	return img
}

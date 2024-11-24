package scene

import (
	image2 "image"
	"image/color"
	"sort"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/yohamta/donburi"

	"github.com/m110/secrets/archetype"
	"github.com/m110/secrets/assets"
	"github.com/m110/secrets/component"
	"github.com/m110/secrets/domain"
)

type debugUIBuilder struct {
	world       donburi.World
	ui          *ebitenui.UI
	buttonImage *widget.ButtonImage
}

func newDebugUIBuilder(world donburi.World) *debugUIBuilder {
	buttonImg := image.NewNineSliceColor(color.RGBA{0, 0, 0, 128})
	hoverImg := image.NewNineSliceColor(color.RGBA{0, 50, 0, 128})

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
			widget.ButtonOpts.Text(levelName, assets.SmallFont, &widget.ButtonTextColor{
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

	dockContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewGridLayout(
			widget.GridLayoutOpts.Columns(1),
			widget.GridLayoutOpts.Spacing(4, 4),
			widget.GridLayoutOpts.Stretch([]bool{true, true, true, true}, []bool{}),
		)),
		widget.ContainerOpts.WidgetOpts(widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
			HorizontalPosition: widget.AnchorLayoutPositionStart,
			VerticalPosition:   widget.AnchorLayoutPositionEnd,
			Padding:            widget.Insets{Left: 50, Bottom: 50},
		})),
	)

	buttons := []struct {
		text    string
		handler func(*widget.ButtonClickedEventArgs)
	}{
		{"Character", func(args *widget.ButtonClickedEventArgs) {
			x, y := characterWindow.Contents.PreferredSize()
			r := image2.Rect(0, 0, x, y)
			r = r.Add(image2.Point{200, 200})
			characterWindow.SetLocation(r)
			b.ui.AddWindow(characterWindow)
		}},
	}

	for _, btn := range buttons {
		button := widget.NewButton(
			widget.ButtonOpts.WidgetOpts(),
			widget.ButtonOpts.Image(b.buttonImage),
			widget.ButtonOpts.Text(btn.text, assets.SmallFont, &widget.ButtonTextColor{
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

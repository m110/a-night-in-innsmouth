package engine

import (
	"image"

	"github.com/yohamta/donburi/features/math"
)

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func NewRect(x, y, width, height float64) Rect {
	return Rect{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

func (r Rect) Position() math.Vec2 {
	return math.Vec2{
		X: r.X,
		Y: r.Y,
	}
}

func (r Rect) Size() Size {
	return Size{
		Width:  int(r.Width),
		Height: int(r.Height),
	}
}

func (r Rect) MaxX() float64 {
	return r.X + r.Width
}

func (r Rect) MaxY() float64 {
	return r.Y + r.Height
}

func (r Rect) Intersects(other Rect) bool {
	return r.X <= other.MaxX() &&
		other.X <= r.MaxX() &&
		r.Y <= other.MaxY() &&
		other.Y <= r.MaxY()
}

func (r Rect) ToImageRectangle() image.Rectangle {
	return image.Rect(
		int(r.X),
		int(r.Y),
		int(r.MaxX()),
		int(r.MaxY()),
	)
}

func (r Rect) Move(pos math.Vec2) Rect {
	return Rect{
		X:      r.X + pos.X,
		Y:      r.Y + pos.Y,
		Width:  r.Width,
		Height: r.Height,
	}
}

func (r Rect) Scale(scale float64) Rect {
	return Rect{
		X:      r.X * scale,
		Y:      r.Y * scale,
		Width:  r.Width * scale,
		Height: r.Height * scale,
	}
}

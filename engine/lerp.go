package engine

import "github.com/yohamta/donburi/features/math"

func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func LerpVec2(a, b math.Vec2, t float64) math.Vec2 {
	return math.Vec2{
		X: Lerp(a.X, b.X, t),
		Y: Lerp(a.Y, b.Y, t),
	}
}

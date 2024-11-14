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

// Ease in - starts slow, ends fast
func EaseIn(t float64) float64 {
	return t * t
}

// Ease out - starts fast, ends slow
func EaseOut(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

// Ease in-out - slow at both ends, smooth in middle
func EaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	t = 2*t - 1
	return -0.5 * (t*(t-2) - 1)
}

// Cubic ease in - even more pronounced slow start
func CubicEaseIn(t float64) float64 {
	return t * t * t
}

// Cubic ease out - even more pronounced slow end
func CubicEaseOut(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}

package engine

func IntPercent(value int, percent float64) int {
	return int(float64(value) * percent)
}

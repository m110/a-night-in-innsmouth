package engine

import (
	"time"
)

type Timer struct {
	currentFrames int
	targetFrames  int
}

func NewTimer(d time.Duration) *Timer {
	t := &Timer{}
	t.SetTarget(d)

	return t
}

func (t *Timer) Update() {
	if t.currentFrames < t.targetFrames {
		t.currentFrames++
	}
}

func (t *Timer) SetTarget(target time.Duration) {
	t.targetFrames = int(target.Milliseconds()) * 60 / 1000
}

func (t *Timer) IsReady() bool {
	return t.currentFrames >= t.targetFrames
}

func (t *Timer) Reset() {
	t.currentFrames = 0
}

func (t *Timer) Finish() {
	t.currentFrames = t.targetFrames
}

func (t *Timer) IsStarted() bool {
	return t.currentFrames > 0
}

func (t *Timer) PercentDone() float64 {
	return float64(t.currentFrames) / float64(t.targetFrames)
}

func (t *Timer) CurrentFrames() int {
	return t.currentFrames
}

func (t *Timer) TargetFrames() int {
	return t.targetFrames
}

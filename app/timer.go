package app

import "time"

type Timer struct {
	lastTick        time.Time
	timeoutDuration time.Duration
}

func NewTimer(timeoutDuration time.Duration) *Timer {
	return &Timer{lastTick: time.Now(), timeoutDuration: timeoutDuration}
}

func (t *Timer) Elapsed() time.Duration {
	return time.Now().Sub(t.lastTick)
}

func (t *Timer) Tick() {
	t.lastTick = time.Now()
}

func (t *Timer) TimedOut() bool {
	return t.Elapsed() > t.timeoutDuration
}

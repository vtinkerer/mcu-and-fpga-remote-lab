package timer

import (
	"fmt"
	"sync"
	"time"
)

type Timer struct {
	duration time.Duration
	timer    *time.Timer
	callback func()
	mu       sync.Mutex
	active   bool
}

func NewTimer(duration time.Duration, callback func()) *Timer {
	return &Timer{
		duration: duration,
	}
}

func (t *Timer) SetDuration(duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.duration = duration
}

func (t *Timer) Start(callback func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timer != nil {
		t.timer.Stop()
	}

	t.active = true
	t.callback = callback
	t.timer = time.AfterFunc(t.duration, func() {
		fmt.Println("Timer expired")
		t.mu.Lock()
		if t.active {
			t.callback()
			t.active = false
		}
		t.mu.Unlock()
	})
}

func (t *Timer) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timer != nil {
		t.timer.Stop()
		t.active = false
	}
}

func (t *Timer) IsActive() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.active
}
package main

import (
	"context"
	"time"
)

// tickInterval is how often the frontend advances elapsed time for the running task.
const tickInterval = 500 * time.Millisecond

// runTimer emits a tick only while a task is running, so an idle app stays asleep.
func (a *App) runTimer(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	if !a.tickerOn.Load() {
		ticker.Stop()
	}
	for {
		select {
		case <-ticker.C:
			emitTick()
		case <-a.tickerWake:
			if a.tickerOn.Load() {
				ticker.Reset(tickInterval)
			} else {
				ticker.Stop()
			}
		case <-ctx.Done():
			return
		}
	}
}

// syncTicker matches the tick to reality: running while a task is active, stopped otherwise.
func (a *App) syncTicker() {
	running, err := GetRunningTask()
	if err != nil {
		a.showError(err)
		return
	}
	a.tickerOn.Store(running != nil)
	select {
	case a.tickerWake <- struct{}{}:
	default:
	}
}

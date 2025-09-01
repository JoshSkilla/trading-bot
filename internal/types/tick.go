package types

import (
	ctx "context"
	"time"
)

type Tick struct {
	Time time.Time
	// Metadata map[string]interface{} // Optional market related information: bear/bull?
}

type TickGenerator func(ctx ctx.Context, ticks chan Tick, runLength int, tickInterval time.Duration)


func NewTick(t time.Time) Tick {
	return Tick{Time: t}
}

// Ticks channel closed outside the following functions:

func GenerateTestTicks(ctx ctx.Context, ticks chan Tick, runLength int, tickInterval time.Duration) {
	for i := range runLength {
		select {
		case <-ctx.Done():
			return
		default:
			ticks <- NewTick(time.Now().Add(time.Duration(i) * tickInterval))
		}
	}
}

func GenerateLiveTicks(ctx ctx.Context, ticks chan Tick, runLength int, tickInterval time.Duration) {
	runDuration := time.Duration(runLength) * tickInterval
	timeOut := time.After(runDuration)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			ticks <- NewTick(now)
		case <-timeOut:
			return
		}
	}
}

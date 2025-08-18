package types

import "time"

type Tick struct {
	Time     time.Time
	// Metadata map[string]interface{} // Optional market related information: bear/bull?

}

func NewTick(t time.Time) Tick {
	return Tick{Time: t}
}

func GenerateTestTicks(ticks chan Tick, runLength int, tickInterval time.Duration) {
	for i := range runLength {
		ticks <- NewTick(time.Now().Add(time.Duration(i) * tickInterval))
	}
}

func GenerateLiveTicks(ticks chan Tick, runLength int, tickInterval time.Duration) {
	runDuration := time.Duration(runLength) * tickInterval
	done := time.After(runDuration)
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			ticks <- NewTick(now)
		case <-done:
			return
		}
	}
}
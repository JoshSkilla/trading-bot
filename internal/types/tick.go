package types

import (
	ctx "context"
	"fmt"
	"time"
)

type Tick struct {
	Time time.Time
	// Metadata map[string]interface{} // Optional market related information: bear/bull?
}

type TickGenerator func(ctx ctx.Context, ticks chan Tick, start time.Time, end time.Time, tickInterval time.Duration, tradingHours TradingHours)

func NewTick(t time.Time) Tick {
	return Tick{Time: t}
}

// Ticks channel closed outside the following functions:

func GenerateTestTicks(ctx ctx.Context, ticks chan Tick, start time.Time, end time.Time, tickInterval time.Duration, th TradingHours) {
	for t := start; t.Before(end); t = t.Add(tickInterval) {
		select {
		case <-ctx.Done():
			return
		default:
			if th.isOpenAt(t) {
				ticks <- NewTick(t)
			} else {
				t = th.getNextOpenTime(t)
			}
		}
	}
}

func GenerateLiveTicks(ctx ctx.Context, ticks chan Tick, start time.Time, end time.Time, tickInterval time.Duration, th TradingHours) {
	// Wait until start
	if now := time.Now(); now.Before(start) {
		timer := time.NewTimer(start.Sub(now))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return
		case <-timer.C:
		}
	}

	for {
		now := time.Now()
		if !now.Before(end) {
			return
		}

		// Wait until market opens
		if !th.isOpenAt(now) {
			nextOpen := th.getNextOpenTime(now)
			if !nextOpen.Before(end) {
				return
			}
			sleep := time.NewTimer(nextOpen.Sub(now))
			select {
			case <-ctx.Done():
				if !sleep.Stop() {
					<-sleep.C
				}
				return
			case <-sleep.C:
			}
		}

		// Market is open
		_, closeTime, err := th.GetTradingHours()
		if err != nil {
			panic(fmt.Errorf("failed to get trading hours: %w", err))
		}

		// Min of end and market closure
		windowEnd := closeTime
		if end.Before(windowEnd) {
			windowEnd = end
		}

		ticker := time.NewTicker(tickInterval)
		windowTimer := time.NewTimer(time.Until(windowEnd))

		// Generates ticks until window closes or context ends
		windowEnded := false
		for !windowEnded {
			select {
			case <-ctx.Done():
				ticker.Stop()
				if !windowTimer.Stop() {
					<-windowTimer.C
				}
				return

			case <-windowTimer.C:
				ticker.Stop()
				windowEnded = true

			case tnow := <-ticker.C:
				ticks <- NewTick(tnow)
			}
		}
	}
}

func shiftToWorkingDay(t time.Time) time.Time {
	if t.Weekday() == time.Saturday {
		t = t.AddDate(0, 0, 2)
	}
	if t.Weekday() == time.Sunday {
		t = t.AddDate(0, 0, 1)
	}
	return t
}

func IntervalStart(ts time.Time, d time.Duration) time.Time {
	ts = ts.UTC()
	sec := int64(d.Seconds())
	return time.Unix((ts.Unix()/sec)*sec, 0).UTC()
}

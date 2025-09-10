package types

import (
	"time"
)

type Aggregator struct {
	Interval time.Duration
	Curr     map[Asset]Bar // bars currently being built
}

func NewAggregator(d time.Duration) *Aggregator {
	return &Aggregator{
		Interval: d,
		Curr:     make(map[Asset]Bar),
	}
}

// PushTrade updates the current bar for the asset, returning a CLOSED bar
// after the interval (otherwise returns nil).
func (a *Aggregator) PushTrade(asset Asset, ts time.Time, price, size float64) *Bar {
	if a.Interval <= 0 {
		// Degenerate trade-as-bar if needed
		// For single trade samples
		p := price
		b := Bar{
			Asset: asset, Start: ts.UTC(), End: ts.UTC(),
			Interval: 0,
			Open:     p, High: p, Low: p, Close: p,
			Volume: size, Notional: p * size, TradeCount: 1, LastTradedVolume: size,
			Status: BarStatusAggregated,
		}
		return &b
	}

	start := IntervalStart(ts, a.Interval)
	end := start.Add(a.Interval)

	b, ok := a.Curr[asset]
	if !ok || !b.Start.Equal(start) {
		// Close previous bar for this asset if any
		if ok && b.Status == BarStatusBuilding {
			b.Status = BarStatusAggregated
			closed := b
			// Start new building bar
			p := price
			a.Curr[asset] = Bar{
				Asset: asset, Start: start, End: end, Interval: a.Interval,
				Open: p, High: p, Low: p, Close: p,
				Volume: size, Notional: p * size, TradeCount: 1, LastTradedVolume: size,
				Status: BarStatusBuilding,
			}
			return &closed
		}
		// No previous: create first building bar
		p := price
		a.Curr[asset] = Bar{
			Asset: asset, Start: start, End: end, Interval: a.Interval,
			Open: p, High: p, Low: p, Close: p,
			Volume: size, Notional: p * size, TradeCount: 1, LastTradedVolume: size,
			Status: BarStatusBuilding,
		}
		return nil
	}

	// Update current building bar
	if price > b.High {
		b.High = price
	}
	if price < b.Low {
		b.Low = price
	}
	b.Close = price
	b.Volume += size
	b.Notional += price * size
	b.TradeCount++
	b.LastTradedVolume = size
	a.Curr[asset] = b
	return nil
}

// Marks all currently building bars as closed and returns them.
func (a *Aggregator) Flush() []Bar {
	out := make([]Bar, 0, len(a.Curr))
	for asset, b := range a.Curr {
		if b.Status == BarStatusBuilding {
			b.Status = BarStatusAggregated
			out = append(out, b)
		}
		delete(a.Curr, asset)
	}
	return out
}

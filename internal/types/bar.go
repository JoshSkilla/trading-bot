package types

import (
	"fmt"
	"time"
)

type BarStatus uint8

const (
	BarStatusBuilding   BarStatus = iota // actively aggregating: not closed yet, not usable
	BarStatusAggregated                  // built from samples, closed, usable (estimations)
	BarStatusOfficial                    // fetched from official bars API (accurate)
	BarStatusNoTrades                    // no trades in this interval (Volume=0, Open=High=Low=Close=last price)
)

// One aggregated period of samples (OHLCV).
type Bar struct {
	Asset    Asset
	Start    time.Time // inclusive start of interval, UTC
	End      time.Time // exclusive end of interval, UTC
	Interval time.Duration

	Open, High, Low, Close float64

	Volume           float64 // sum of size of all trades in interval
	LastTradedVolume float64 // size of last trade in interval
	Notional         float64 // sum(price*size). Used in: VWAP = Notional/Volume
	TradeCount       int

	Status BarStatus
}

// --- Constructors & Mutators ---

// Create new bar from single sample with given interval
func NewBarFromSampleWithInterval(s Sample, start, end time.Time) Bar {
	return Bar{
		Asset:            s.Asset,
		Start:            start,
		End:              end,
		Interval:         end.Sub(start),
		Open:             s.Price,
		High:             s.Price,
		Low:              s.Price,
		Close:            s.Price,
		Volume:           s.Volume,
		Notional:         s.Price * s.Volume,
		TradeCount:       1,
		LastTradedVolume: s.Volume,
		Status:           BarStatusAggregated,
	}
}

func NewCarryForwardBarFromSample(s Sample, start, end time.Time) Bar {
	return Bar{
		Asset:            s.Asset,
		Start:            start,
		End:              end,
		Interval:         end.Sub(start),
		Open:             s.Price,
		High:             s.Price,
		Low:              s.Price,
		Close:            s.Price,
		Volume:           0,
		Notional:         0,
		TradeCount:       0,
		LastTradedVolume: s.Volume,
		Status:           BarStatusNoTrades,
	}
}

// Create new bar from single sample with zero interval
func NewBarFromSample(s Sample) Bar {
	return NewBarFromSampleWithInterval(s, s.Time.UTC(), s.Time.UTC())
}

// UpdateWithTrade mutates b with a single trade (ts, price, size).
func (b *Bar) UpdateWithTrade(ts time.Time, price, size float64) bool {
	if !b.InRange(ts) {
		return false
	}
	if b.TradeCount == 0 {
		b.Open, b.High, b.Low, b.Close = price, price, price, price
	} else {
		if price > b.High {
			b.High = price
		}
		if price < b.Low {
			b.Low = price
		}
		b.Close = price
	}
	b.Volume += size
	b.Notional += price * size
	b.TradeCount++
	b.LastTradedVolume = size

	if b.Status == BarStatusNoTrades {
		b.Status = BarStatusAggregated
	}

	return true
}

// UpdateWithSample mutates b with a single sample s.
func (b *Bar) UpdateWithSample(s Sample) bool {
	return b.UpdateWithTrade(s.Time, s.Price, s.Volume)
}

// --- Queries & Formatting---

// VWAP returns the Volume Weighted Average Price for the bar.
// If Volume is zero, VWAP is defined to be zero.
func (b Bar) VWAP() float64 {
	if b.Volume == 0 {
		return 0
	}
	return b.Notional / b.Volume
}

// InRange returns true if ts ∈ [Start, End).
func (b Bar) InRange(ts time.Time) bool {
	tu := ts.UTC()
	return !tu.Before(b.Start) && tu.Before(b.End)
}

// Pretty returns a formatted string representation of the Bar.
func (b Bar) Pretty() string {
	const layout = "2006-01-02 15:04:05 MST"
	return fmt.Sprintf(
		"%s [%s - %s] (%s)\nOHLC: %.2f %.2f %.2f %.2f  Volume: %.2f",
		b.Asset.Symbol,
		b.Start.Format(layout),
		b.End.Format(layout),
		b.Interval,
		b.Open, b.High, b.Low, b.Close, b.Volume,
	)
}

func (b Bar) IsClosed() bool      { return b.Status != BarStatusBuilding }
func (b Bar) IsOfficial() bool    { return b.Status == BarStatusOfficial }
func (b Bar) IsNoTrades() bool    { return b.Status == BarStatusNoTrades }
func (b Bar) IsThin(min int) bool { return b.TradeCount < min }
func (b Bar) IsTickOnly() bool    { return b.TradeCount == 1 && b.Interval == 0 }

func (b *Bar) SetStatus(s BarStatus) {
	b.Status = s
}
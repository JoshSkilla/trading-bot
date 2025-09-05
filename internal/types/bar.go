package types

import (
	"time"
)

// One aggregated period of samples (OHLCV).
type Bar struct {
	Asset    Asset
	Start    time.Time // start of interval
	End      time.Time // end of interval
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
	Interval time.Duration
}

func NewEmptyBar(asset Asset, start, end time.Time) Bar {
	return Bar{
		Asset:    asset,
		Start:    start,
		End:      end,
		Interval: end.Sub(start),
	}
}

func NewBarFromSample(s Sample, start, end time.Time) Bar {
	return Bar{
		Asset:    s.Asset,
		Start:    start,
		End:      end,
		Interval: end.Sub(start),
		Open:     s.Price,
		High:     s.Price,
		Low:      s.Price,
		Close:    s.Price,
		Volume:   s.Volume,
	}
}


func SamplesToBar(asset Asset, samples []Sample, start, end time.Time) Bar {
	bar := NewEmptyBar(asset, start, end)

	if len(samples) == 0 {
		return bar
	}

	// Initialize bar from the first sample
	bar = NewBarFromSample(samples[0], start, end)

	// Process the rest with AddSample
	for _, s := range samples[1:] {
		bar.AddSample(s)
	}

	return bar
}

func (b *Bar) AddSample(s Sample) {
	if b.Open == 0 && b.High == 0 && b.Low == 0 && b.Close == 0 && b.Volume == 0 {
		b.Open = s.Price
		b.High = s.Price
		b.Low  = s.Price
	}
	if s.Price > b.High {
		b.High = s.Price
	}
	if s.Price < b.Low {
		b.Low = s.Price
	}
	b.Close  = s.Price
	b.Volume += s.Volume
}
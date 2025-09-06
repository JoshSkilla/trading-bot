package types

import (
	"fmt"
	"time"
)

// One instantaneous market observation.
type Sample struct {
	Asset  Asset
	Time   time.Time
	Price  float64
	Volume float64
}

func NewSample(asset Asset, t time.Time, price, volume float64) Sample {
	return Sample{
		Asset:  asset,
		Time:   t,
		Price:  price,
		Volume: volume,
	}
}

// Pretty returns a human-readable string representation of the sample.
func (s Sample) Pretty() string {
	return fmt.Sprintf(
		"%s @ %s | Price: %.2f | Volume: %.2f",
		s.Asset.Symbol,
		s.Time.Format("2006-01-02 15:04:05 MST"),
		s.Price,
		s.Volume,
	)
}

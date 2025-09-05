package types

import "time"

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
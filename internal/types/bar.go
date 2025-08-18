package types

import (
	"time"
)

type Bar struct {
	Asset  Asset
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}
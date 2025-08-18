package types

import (
	"time"

)

// Signal is emitted by a strategy when it wants to act
type Signal struct {
	Time       time.Time
	Bar        Bar
	Action     Action
	Qty        float64
	Confidence float64
}

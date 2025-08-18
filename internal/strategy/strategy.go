package strategy

import (
	t "github.com/joshskilla/trading-bot/internal/types"
)


// Strategy interface is implemented by any trading strategy
type Strategy interface {
	Init() error
	OnTick(tick t.Tick) // Update internal state
	GenerateSignals() []t.Signal
	Name() string
}

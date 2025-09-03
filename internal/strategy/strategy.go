package strategy

import (
	t "github.com/joshskilla/trading-bot/internal/types"

	"fmt"
	"time"
)

// Strategy interface is implemented by any trading strategy
type Strategy interface {
	Init() error
	OnTick(tick t.Tick) // Update internal state
	GenerateSignals() []t.Signal
	Name() string
	TickInterval() time.Duration
}

func RestoreFromCheckpoint(strategyType string, checkpoint *Checkpoint) (Strategy, error) {
	var strat Strategy
	var err error = nil
	switch strategyType {
	case "momentum":
		strat, err = RestoreMomentumStrategy(checkpoint)
	default:
		return nil, fmt.Errorf("unknown strategy type: %s", strategyType)
	}
	return strat, err
}

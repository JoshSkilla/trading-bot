package strategy

import (
	"time"

	t "github.com/joshskilla/trading-bot/internal/types"
)

type MomentumStrategy struct {
	asset  t.Asset
	bar    t.Bar
	fastMA float64
	slowMA float64
}

func NewMomentumStrategy(asset t.Asset) *MomentumStrategy {
	return &MomentumStrategy{asset: asset, bar: t.Bar{}, fastMA: 0, slowMA: 0}
}

func (m *MomentumStrategy) Init() error {
	return nil
}

func (m *MomentumStrategy) OnTick(tick t.Tick) {
	// get updated asset information (bar)
	// update fast & slow MA
}

func (m *MomentumStrategy) GenerateSignals() []t.Signal {
	// check for fast & slow crosses

	// if want to smooth out buying process make this apparent in generate signals
	// (perhaps use shared functions that handle this)
	// also consider the amount spent per signal - may need another attribute

	// placeholder
	return []t.Signal{
		{
			Time:       time.Now(),
			Bar:        m.bar,
			Action:     t.Buy,
			Confidence: 0.9,
		},
	}
}

func (m *MomentumStrategy) Name() string {
	return "Momentum"
}

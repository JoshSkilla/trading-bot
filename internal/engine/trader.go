package engine

import (
	"time"

	t "github.com/joshskilla/trading-bot/internal/types"
)

type Trader interface {
	Execute(*Portfolio, t.Signal) (ExecutionRecord, bool)
}

type LiveTrader struct{}

type PaperTrader struct{}

type TestTrader struct{}

func NewTestTrader() *TestTrader {
	return &TestTrader{}
}

func (tt *TestTrader) Execute(p *Portfolio, sig t.Signal) (ExecutionRecord, bool) {
	qty := sig.Qty
	price := sig.Bar.Close
	asset := sig.Bar.Asset
	// live execute will only add it to execHistory once order fulfilled - and get price then
	cost := qty * price
	if sig.Action == t.Buy && p.Cash >= cost {
		p.Cash -= cost
		p.Positions[asset] += qty
		exec := ExecutionRecord{time.Now(), asset, t.Buy, qty, price, p.Cash}
		return exec, true
	}
	return ExecutionRecord{}, false
}

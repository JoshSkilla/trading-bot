package engine

import (
	st "github.com/joshskilla/trading-bot/internal/strategy"
	t "github.com/joshskilla/trading-bot/internal/types"
)

type Runner struct {
	Portfolio *Portfolio
	Trader    Trader
	Strategy  st.Strategy
	Ticks     chan t.Tick
}

const MaxExecutionHistory = 10

func NewRunner(p *Portfolio, t Trader, s st.Strategy, ch chan t.Tick) *Runner {
	return &Runner{
		Portfolio: p,
		Trader:    t,
		Strategy:  s,
		Ticks:     ch,
	}
}

func (r *Runner) Run() {
	for t := range r.Ticks {
		r.Strategy.OnTick(t)
		for _, sig := range r.Strategy.GenerateSignals() {
			execRecord, ok := r.Trader.Execute(r.Portfolio, sig)
			if !ok {
				continue
			}
			if len(r.Portfolio.ExecutionHistory) >= MaxExecutionHistory {
				r.Portfolio.FlushToCSV(true)
			}
			r.Portfolio.ExecutionHistory = append(r.Portfolio.ExecutionHistory, execRecord)
		}
	}
}
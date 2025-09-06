package engine

import (
	ctx "context"
	"fmt"

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

func (r *Runner) Run(ctx ctx.Context) {
	// Cleanup on exit
	defer func() {
		r.Trader.Close() // ensure trader resources are cleaned up
		fmt.Printf("Trader closed for strategy %s on portfolio %s\n", r.Strategy.Name(), r.Portfolio.Name)
	}()

	for {
		select {
		case <-ctx.Done():
			// Runner cancelled:
			// Flush remaining executions before exit
			// (Cancel live orders TODO)
			// Stop the runner
			if len(r.Portfolio.ExecutionHistory) > 0 {
				r.Portfolio.FlushOrdersToFile()
			}
			fmt.Printf("Shut down strategy %s on portfolio %s...\n", r.Strategy.Name(), r.Portfolio.Name)
			return
		case t, ok := <-r.Ticks:
			if !ok {
				// Completed ticks (channel closed):
				// (Wait on live orders TODO)
				if len(r.Portfolio.ExecutionHistory) > 0 {
					r.Portfolio.FlushOrdersToFile()
				}
				fmt.Printf("Finished processing for strategy %s on portfolio %s...\n", r.Strategy.Name(), r.Portfolio.Name)
				return
			}
			r.Strategy.OnTick(t)
			for _, sig := range r.Strategy.GenerateSignals() {
				execRecord, ok := r.Trader.Execute(r.Portfolio, sig)
				if !ok {
					continue
				}
				r.Portfolio.ExecutionHistory = append(r.Portfolio.ExecutionHistory, execRecord)
				if len(r.Portfolio.ExecutionHistory) >= MaxExecutionHistory {
					r.Portfolio.FlushOrdersToFile()
				}
				r.Portfolio.FlushPositionsToFile()
			}
		}
	}
}

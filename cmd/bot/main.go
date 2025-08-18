package main

import (
	"sync"
	"time"

	"github.com/joshskilla/trading-bot/internal/engine"
	st "github.com/joshskilla/trading-bot/internal/strategy"
	t "github.com/joshskilla/trading-bot/internal/types"
)

func main() {
	cash := 100000.0
	name := "test"
	asset := t.NewAsset("AAPL", "NASDAQ", "stock")
	portfolio := engine.NewPortfolio(name, cash)
	strat := st.NewMomentumStrategy(asset)
	trader := engine.NewTestTrader()
	ticks := make(chan t.Tick, 10)
	runner := engine.NewRunner(portfolio, trader, strat, ticks)

	isTest := true
	tickInterval := time.Hour
	runLength := 100

	var wg sync.WaitGroup

	go func() {
		defer close(ticks)
		if isTest {
			t.GenerateTestTicks(ticks, runLength, tickInterval)
		} else {
			t.GenerateLiveTicks(ticks, runLength, tickInterval)
		}
	}()

	wg.Go(func() {
		runner.Run()
	})
	wg.Wait()
}
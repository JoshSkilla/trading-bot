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

// could have a while loop to receive terminal instructions commanding go routines
// ie telling runner(s) to shutdown and cancel orders cleanly
// strategies should have tests for suitability across a period
// strategies might need a fitting function to fit internals
// update trader execute - for simulated trading
// file struct and asset struct, asset and market and label etc?
// decide on signal processing - unarity or external functions can be applied like smoothing
// alpaca adpator and finish trader execute functinos
// do flush csv methods
// Actually code momentum
// save runners in json so they can be reloaded

package main

import (
	"bufio"
	ctx "context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joshskilla/trading-bot/internal/engine"
	st "github.com/joshskilla/trading-bot/internal/strategy"
	t "github.com/joshskilla/trading-bot/internal/types"
)

func main() {
	// Create runners & portfolios
	cash := 100000.0
	name := "test"
	asset := t.NewAsset("AAPL", "NASDAQ", "stock")
	portfolio := engine.NewPortfolio(name, cash)
	strat := st.NewMomentumStrategy(asset)
	trader := engine.NewTestTrader()
	ticks := make(chan t.Tick, 10)
	runner := engine.NewRunner(portfolio, trader, strat, ticks)

	// Setup
	var wg sync.WaitGroup
	ctx, cancel := ctx.WithCancel(ctx.Background())
	cmdChan := make(chan string)
	isTest := true
	tickInterval := time.Hour
	runLength := 100

	var tickGen t.TickGenerator
	if isTest {
		tickGen = t.GenerateTestTicks
	} else {
		tickGen = t.GenerateLiveTicks
	}

	// Generate ticks for runner(s)
	go func() {
		defer close(ticks)
		tickGen(ctx, ticks, runLength, tickInterval)
	}()

	// Run runner(s)
	wg.Go(func() {
		runner.Run(ctx)
	})

	// Command line input goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := scanner.Text()
			cmdChan <- cmd
		}
		// Handle scanner.Err()?
	}()

	// Main loop: handle live commands
	for {
		select {
		case cmd := <-cmdChan:
			if cmd == "--shutdown" {
				cancel() // Signal runners to stop
				fmt.Println("Shutting down trading-bot...")
				return
			}
			// Other commands...

		case <-ctx.Done():
			// Wait on runners to gracefully shut down
			wg.Wait()
			return
		}
	}
}

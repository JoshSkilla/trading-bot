package engine

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	st "github.com/joshskilla/trading-bot/internal/strategy"
	t "github.com/joshskilla/trading-bot/internal/types"
	cfg "github.com/joshskilla/trading-bot/config"
)

// Runs the trading session
// Coordinates the runners, tick generators, trader, and live command inputs
func Run(portfolio *Portfolio, strat st.Strategy, trader *TestTrader, isTest bool, start time.Time, end time.Time) error {

	ticks := make(chan t.Tick, 10)
	tickInterval := strat.TickInterval()
	runner := NewRunner(portfolio, trader, strat, ticks)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	cmdChan := make(chan string)

	var tickGen t.TickGenerator
	if isTest {
		tickGen = t.GenerateTestTicks
	} else {
		tickGen = t.GenerateLiveTicks
	}

	tradingHours := t.TradingHours{
		OpenHour:    cfg.OpenHour,
		OpenMinute:  cfg.OpenMinute,
		CloseHour:   cfg.ClosingHour,
		CloseMinute: cfg.ClosingMinute,
		WeekendsOff: true,
		ExchangeTZ:  cfg.ExchangeTimeZone,
	}

	// Generate ticks for runner(s)
	go func() {
		defer close(ticks)
		tickGen(ctx, ticks, start, end, tickInterval, tradingHours)
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
	}()

	// Main loop: handle live commands
	for {
		select {
		case cmd := <-cmdChan:
			if cmd == "--shutdown" {
				cancel() // Signal runners to stop
				fmt.Println("Shutting down trading-bot...")
				return nil
			}
			// Other commands...
		case <-ctx.Done():
			// Wait on runners to gracefully shut down
			wg.Wait()
			cancel() // Not necessary as its a given
			fmt.Println("All runners shut down. Exiting.")
			return nil
		}
	}
}
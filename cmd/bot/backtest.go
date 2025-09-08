package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshskilla/trading-bot/internal/engine"
	st "github.com/joshskilla/trading-bot/internal/strategy"
	"github.com/urfave/cli/v3"
)

func BacktestCmd() *cli.Command {
	return &cli.Command{
		Name:  "backtest",
		Usage: "Test a portfolio with a strategy & checkpoint over a period",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "portfolio", Aliases: []string{"p"}, Usage: "Portfolio name", Required: true},
			&cli.StringFlag{Name: "strategy", Aliases: []string{"s"}, Usage: "Strategy name", Required: true},
			&cli.StringFlag{Name: "checkpoint", Aliases: []string{"c"}, Usage: "Checkpoint label or id"},
			&cli.StringFlag{Name: "start", Aliases: []string{"s"}, Usage: "Start time for backtest", Required: true},
			&cli.StringFlag{Name: "end", Aliases: []string{"e"}, Usage: "End time for backtest", Required: true},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			portfolioName := c.String("portfolio")
			strategyType := c.String("strategy")
			checkpointName := c.String("checkpoint")
			startStr := c.String("start")
			endStr := c.String("end")

			if portfolioName == "" || strategyType == "" || checkpointName == "" || startStr == "" || endStr == "" {
				return errors.New("--portfolio, --strategy, --checkpoint, --start, and --end are required")
			}

			var start, end time.Time
			var err error
			if startStr != "" {
				start, err = time.Parse(time.RFC3339, startStr)
				if err != nil {
					return fmt.Errorf("invalid start time: %w", err)
				}
			}
			if endStr != "" {
				end, err = time.Parse(time.RFC3339, endStr)
				if err != nil {
					return fmt.Errorf("invalid end time: %w", err)
				}
			}

			// Load portfolio and checkpoint
			fmt.Printf("Restoring portfolio %q with strategy %q from checkpoint %q\n", portfolioName, strategyType, checkpointName)
			portfolio, err := engine.LoadPortfolioFromJSON(portfolioName)
			if err != nil {
				return fmt.Errorf("failed to load portfolio: %w", err)
			}

			var checkpoint *st.Checkpoint
			checkpoint, err = st.LoadCheckpointFromJSON(checkpointName)
			if err != nil {
				return fmt.Errorf("failed to get checkpoint: %w", err)
			}

			// Restore strategy state
			var strat st.Strategy
			strat, err = st.RestoreFromCheckpoint(strategyType, checkpoint)
			if err != nil {
				return fmt.Errorf("failed to restore strategy from checkpoint: %w", err)
			}

			trader := engine.NewTestTrader(strat.TickInterval(), start, end)
			// Run the trading session
			return engine.Run(portfolio, strat, trader, true, start, end)
		},
	}
}

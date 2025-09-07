package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshskilla/trading-bot/internal/engine"
	st "github.com/joshskilla/trading-bot/internal/strategy"
	cfg "github.com/joshskilla/trading-bot/internal/config"

	"github.com/urfave/cli/v3"
)

func RunCmd() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run a portfolio with a strategy & checkpoint",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "portfolio", Aliases: []string{"p"}, Usage: "Portfolio name", Required: true},
			&cli.StringFlag{Name: "strategy", Aliases: []string{"s"}, Usage: "Strategy name", Required: true},
			&cli.StringFlag{Name: "checkpoint", Aliases: []string{"c"}, Usage: "Checkpoint label or id"},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			portfolioName := c.String("portfolio")
			strategyType := c.String("strategy")
			checkpointName := c.String("checkpoint")

			if portfolioName == "" || strategyType == "" || checkpointName == "" {
				return errors.New("--portfolio, --strategy, and --checkpoint are required")
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

			trader := engine.NewTestTrader() // replace with live trader once functional

			// Run the trading session
			start := time.Now()
			defaultEnd := start.Add(cfg.MaxLiveTradingDuration)
			return engine.Run(portfolio, strat, trader, false, start, defaultEnd)
		},
	}
}

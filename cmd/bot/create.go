package main

import (
	"context"
	"fmt"
	"github.com/urfave/cli/v3"
	"strconv"

	t "github.com/joshskilla/trading-bot/internal/types"
	"github.com/joshskilla/trading-bot/internal/engine"
	st "github.com/joshskilla/trading-bot/internal/strategy"
	cfg "github.com/joshskilla/trading-bot/config"
)

// Create resources (e.g., portfolios, checkpoints)
// USAGE: bot create portfolio --name "My Portfolio" --cash 1000
func CreateCmd() *cli.Command {
  return &cli.Command{
	Name:  "create",
	Usage: "Create resources (e.g., portfolios, checkpoints)",
	Commands: []*cli.Command{
		{
			Name:  "portfolio",
			Usage: "Create a new portfolio",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Portfolio name", Required: true},
				&cli.Float64Flag{Name: "cash", Aliases: []string{"c"}, Usage: "Starting cash", Required: true},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				name := c.String("name")
				cash := c.Float64("cash")

				engine.NewPortfolio(name, cash).SaveToJSON()
				fmt.Printf("Created portfolio %q with cash=%.2f \n", name, cash)
				return nil
			},
		},
		{
			Name:  "checkpoint",
			Usage: "Create a new checkpoint",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "checkpoint id",
					Required: true,
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				id := c.String("id")

				cp, err := ParseArgsForCheckpoint(id, c.Args().Slice())
				if err != nil {
					return fmt.Errorf("failed to create checkpoint: %w", err)
				}
				fmt.Printf("Created checkpoint %q\n", cp.ID)
				cp.SaveToJSON()
				return nil
			},
		},
	},
  }
}

// Limited functionality should not be relied upon, numbers only to float64
func ParseArgsForCheckpoint(id string, args []string) (*st.Checkpoint, error) {
	kv, err := st.ParseKV(args)
	if err != nil {
		return nil, err
	}
	attributes := make(map[string]any)
	for k, v := range kv {
		if k == "asset" {
			// Restricted to one assetType and exchange for now
			attributes[k] = t.Asset{
				Symbol:   v,
				Type:     cfg.AssetType,
				Exchange: cfg.Exchange,
			}
		}
		// Convert numeric strings to float64
		if len(v) > 0 && (v[0] >= '0' && v[0] <= '9') {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				attributes[k] = f
			}
		}
	}
	cp := &st.Checkpoint{
		ID:         id,
		Attributes: attributes,
	}
	return cp, nil
}
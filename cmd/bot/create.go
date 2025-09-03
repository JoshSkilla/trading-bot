package main

import (
  "context"
  "fmt"
  "github.com/urfave/cli/v3"
  "github.com/joshskilla/trading-bot/internal/engine"
  st "github.com/joshskilla/trading-bot/internal/strategy"
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

				cp, err := st.ParseArgsForCheckpoint(id, c.Args().Slice())
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
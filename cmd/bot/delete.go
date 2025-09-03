package main

import (
  "context"
  "fmt"
  "github.com/urfave/cli/v3"
  "os"

  e "github.com/joshskilla/trading-bot/internal/engine"
  st "github.com/joshskilla/trading-bot/internal/strategy"
)

// Delete resources (e.g., portfolios, checkpoints)
// USAGE: bot delete portfolio --name "MyPortfolio"
func DeleteCmd() *cli.Command {
  return &cli.Command{
	Name:  "delete",
	Usage: "Delete resources (e.g., portfolios, checkpoints)",
	Commands: []*cli.Command{
		{
			Name:  "portfolio",
			Usage: "Delete a portfolio",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "name",
					Usage:    "portfolio name",
					Required: true,
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				name := c.String("name")

				err := os.Remove(fmt.Sprintf(e.PortfolioFilePath, name))
				if err != nil && !os.IsNotExist(err) {
					return err
				}

				err = os.Remove(fmt.Sprintf(e.OrdersFilePath, name))
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				err = os.Remove(fmt.Sprintf(e.PositionsFilePath, name))
				if err != nil && !os.IsNotExist(err) {
					return err
				}

				fmt.Printf("Deleted portfolio %q and its result CSVs \n", name)
				return nil
			},
		},
		{
			Name:  "checkpoint",
			Usage: "Delete a checkpoint",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "checkpoint id",
					Required: true,
				},
			},
			Action: func(ctx context.Context, c *cli.Command) error {
				id := c.String("id")

				err := os.Remove(fmt.Sprintf(st.CheckpointFilePath, id))
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				fmt.Printf("Deleted checkpoint %q\n", id)
				return nil
			},
		},
	},
  }
}
package main

import (
	"context"
	"fmt"
	"os"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			CreateCmd(), 
			DeleteCmd(),
			RunCmd(),
			BacktestCmd(),
		},
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println("Error running bot:", err)
	}
}

package main

import (
	"github.com/urfave/cli/v2"
)

func RunCli() *cli.App {
	return &cli.App{
		Name:  "Docknat",
		Usage: "Custom utility in Go that updates NAT table in ip-tables based on docker container changes on interfaces",
		Action: func(c *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Start the docknat service",
				Action: func(c *cli.Context) error {
					StartDocknat()
					return nil
				},
			},
		},
	}

}

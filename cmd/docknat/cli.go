package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func RunCli() *cli.App {
	config := Config{
		path: "/etc/docknat/docknat.yaml",
	}
	return &cli.App{
		Name:  "Docknat",
		Usage: "Custom utility in Go that updates NAT table in ip-tables based on docker container changes on interfaces",
		Action: func(c *cli.Context) error {
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the docknat service",
				Action: func(c *cli.Context) error {
					StartDocknat(config)
					return nil
				},
			},
			{
				Name:  "stop",
				Usage: "Stop the docknat service",
				Action: func(c *cli.Context) error {
					fmt.Println("Stopping docknat service")
					return nil
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "Load configuration from `FILE`",
				DefaultText: "/etc/docknat/docknat.yaml",
				Action: func(c *cli.Context, value string) error {
					// Add the configuration file to the config struct
					config.path = value
					return nil
				},
			},
		},
	}

}

package main

import (
	"fmt"
	"os"

	"github.com/coreos/go-iptables/iptables"
	utils "github.com/lukaszmoskwa/docknat/internal"
	"github.com/urfave/cli/v2"
)

type Config struct {
	// Configuration for docknat
	path string
	// Configuration for iptables
	iptables struct {
		// The name of the chain to use
		chain string
		// The name of the table to use
		table string
	}
	// Configuration for docker
	docker struct {
		// The name of the docker network to use
		network string
	}
}

func StartDocknat(config Config) {
	fmt.Println("Starting docknat service")
	fmt.Println("Configuration file:", config.path)
	ipt, error := iptables.New()
	if error != nil {
		fmt.Println("Error creating iptables object: (", error, ")")
		panic(error)
	}
	l := utils.RetrieveNatRules(ipt)
	fmt.Println(l)
	Run()
}

func main() {
	config := Config{
		path: "/etc/docknat/docknat.yaml",
	}
	app := &cli.App{
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

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

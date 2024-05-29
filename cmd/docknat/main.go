package main

import (
	"fmt"
	"os"
)

var (
	Version string
)

type Config struct {
	// Configuration for docknat
	path string
	// Configuration for iptables
}

func StartDocknat(config Config) {
	fmt.Println("Starting docknat service")
	fmt.Println("Configuration file:", config.path)
	fmt.Println("Version:", string(Version))
	Run()
}

func main() {
	app := RunCli()
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"os"
)

type Config struct {
	// Configuration for docknat
	path string
	// Configuration for iptables
}

func StartDocknat(config Config) {
	fmt.Println("Starting docknat service")
	fmt.Println("Configuration file:", config.path)
	Run()
}

func main() {
	app := RunCli()
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

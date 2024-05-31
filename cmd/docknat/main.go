package main

import (
	"fmt"
	"os"
)

var (
	Version string
)

func StartDocknat() {
	fmt.Println("Starting docknat service")
	fmt.Println("Version:", string(Version))
	Run()
}

func main() {
	app := RunCli()
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

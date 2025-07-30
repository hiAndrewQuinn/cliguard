package main

import (
	"os"

	"github.com/cliguard/test/subcommands/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
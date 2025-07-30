package main

import (
	"os"

	"github.com/cliguard/test/breaking-v1/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
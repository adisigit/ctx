package main

import (
	"ctx/internal/cli"
	"fmt"
	"os"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
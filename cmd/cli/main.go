package main

import (
	"fmt"
	"os"

	"github.com/raykavin/docchain/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd()
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
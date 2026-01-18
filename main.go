// Package main is the entry point for the MTC (Merkle Tree Checksum) CLI application.
// It initializes all subcommands and executes the root command.
package main

import (
	"github.com/lucho00cuba/mtc/cmd"
	_ "github.com/lucho00cuba/mtc/cmd/calc"
	_ "github.com/lucho00cuba/mtc/cmd/diff"
	_ "github.com/lucho00cuba/mtc/cmd/hash"
)

// main is the entry point of the application.
// It executes the root command which handles all CLI interactions.
func main() {
	cmd.Execute()
}

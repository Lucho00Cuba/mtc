// Package diff provides the "diff" command for comparing two directory trees
// by computing their Merkle root hashes and reporting differences.
package diff

import (
	"fmt"
	"time"

	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/lucho00cuba/mtc/internal/merkle"

	"github.com/lucho00cuba/mtc/cmd"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command for directory comparison.
var diffCmd = &cobra.Command{
	Use:   "diff [pathA] [pathB]",
	Short: "Compare two directory Merkle trees",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		pathA := args[0]
		pathB := args[1]
		log := logger.With("pathA", pathA, "pathB", pathB, "command", "diff")

		// Read flags directly from command to ensure they're parsed correctly
		patterns, err := cmd.Flags().GetStringArray("exclude")
		if err != nil {
			log.Warn("Failed to read exclude patterns", "error", err)
			patterns = []string{}
		}
		customIgnoreFile, err := cmd.Flags().GetString("ignore-file")
		if err != nil {
			log.Warn("Failed to read ignore-file flag", "error", err)
			customIgnoreFile = ""
		}

		log.Info("Starting directory comparison")
		start := time.Now()

		diff, err := merkle.CompareWithExclusions(pathA, pathB, patterns, true, customIgnoreFile)
		if err != nil {
			log.Error("Comparison failed", "error", err, "duration", time.Since(start))
			return err
		}

		duration := time.Since(start)
		log.Info("Comparison completed",
			"duration", duration,
			"differences", len(diff),
		)

		// Output to stdout (for piping)
		for _, d := range diff {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), d); err != nil {
				log.Error("Failed to write output to stdout", "error", err, "line", d)
				return fmt.Errorf("failed to write output: %w", err)
			}
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().StringArrayP("exclude", "e", []string{}, "Exclude patterns (e.g., 'node_modules', '.git'). Can be specified multiple times.")
	diffCmd.Flags().StringP("ignore-file", "i", "", "Path to a custom ignore file (takes highest priority). .mtcignore and .gitignore are always loaded automatically from the working directory.")

	cmd.Register(diffCmd)
}

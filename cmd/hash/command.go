// Package hash provides the "hash" command for computing Merkle root hashes
// of files and directories. This is the primary command for generating checksums.
package hash

import (
	"fmt"
	"os"
	"time"

	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/lucho00cuba/mtc/internal/merkle"

	"github.com/lucho00cuba/mtc/cmd"
	"github.com/spf13/cobra"
)

// hashCmd represents the hash command for computing Merkle root hashes.
var hashCmd = &cobra.Command{
	Use:   "hash [path]",
	Short: "Compute Merkle root hash of a file or directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		log := logger.With("path", path, "command", "hash")

		// Read flags directly from command to ensure they're parsed correctly
		excludePatterns, err := cmd.Flags().GetStringArray("exclude")
		if err != nil {
			log.Warn("Failed to read exclude patterns", "error", err)
			excludePatterns = []string{}
		}
		customIgnoreFile, err := cmd.Flags().GetString("ignore-file")
		if err != nil {
			log.Warn("Failed to read ignore-file flag", "error", err)
			customIgnoreFile = ""
		}

		log.Info("Starting hash computation")
		start := time.Now()

		// Get path info once to determine type for output
		pathInfo, err := os.Stat(path)
		if err != nil {
			log.Error("Failed to get path info", "error", err)
			return fmt.Errorf("failed to stat path %q: %w", path, err)
		}

		isDir := pathInfo.IsDir()

		// Always create engine with exclusions (automatically loads .mtcignore and .gitignore)
		// Custom ignore file and exclude patterns are optional additions
		engine, err := merkle.NewEngineWithExclusions(0, excludePatterns, path, true, customIgnoreFile)
		if err != nil {
			log.Error("Failed to create engine with exclusions", "error", err)
			return fmt.Errorf("failed to create engine: %w", err)
		}
		result, err := engine.HashPath(path)
		if err != nil {
			log.Error("Hash computation failed", "error", err, "duration", time.Since(start))
			return err
		}

		duration := time.Since(start)
		log.Info("Hash computation completed",
			"duration", duration,
			"hash", fmt.Sprintf("%x", result.Hash),
			"size", formatSize(result.Size),
		)

		// Output to stdout (for piping)
		pathType := "f"
		if isDir {
			pathType = "d"
		}
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (%s): %x (size: %s)\n",
			path, pathType, result.Hash, formatSize(result.Size)); err != nil {
			log.Error("Failed to write output to stdout", "error", err)
			return fmt.Errorf("failed to write output: %w", err)
		}
		return nil
	},
}

// formatSize formats a size in bytes to a human-readable string.
// It automatically selects the most appropriate unit (B, KB, MB, GB, TB, PB, EB)
// based on the size value. Uses binary (1024-based) units.
//
// The function uses 1 decimal place for MB and above, and shows integers for KB
// when the decimal part is zero.
//
// Parameters:
//   - bytes: The size in bytes to format
//
// Returns a formatted string like "1.5 MB" or "512 B".
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	size := float64(bytes)
	exp := 0

	for size >= unit && exp < len(units)-1 {
		size /= unit
		exp++
	}

	// Use 1 decimal place for MB and above, but for KB show as integer if decimal is zero
	if exp == 1 { // KB
		if size == float64(int64(size)) {
			return fmt.Sprintf("%.0f %s", size, units[exp])
		}
		return fmt.Sprintf("%.1f %s", size, units[exp])
	}
	// For MB and above, always show 1 decimal place
	return fmt.Sprintf("%.1f %s", size, units[exp])
}

func init() {
	hashCmd.Flags().StringArrayP("exclude", "e", []string{}, "Exclude patterns (e.g., 'node_modules', '.git'). Can be specified multiple times.")
	hashCmd.Flags().StringP("ignore-file", "i", "", "Path to a custom ignore file (takes highest priority). .mtcignore and .gitignore are always loaded automatically from the working directory.")

	cmd.Register(hashCmd)
}

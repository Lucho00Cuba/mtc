// Package calc provides the "calc" command for verifying that a file or directory
// matches a given Merkle root hash. This is useful for integrity verification.
package calc

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/lucho00cuba/mtc/internal/merkle"

	"github.com/lucho00cuba/mtc/cmd"
	"github.com/spf13/cobra"
)

// calcCmd represents the calc command for hash verification.
var calcCmd = &cobra.Command{
	Use:   "calc [path] [hash]",
	Short: "Verify that a file or directory matches the given hash",
	Long: `Verify that a file or directory matches the given hash.
Computes the Merkle root hash of the specified path and compares it with the provided hash.
Exits with code 0 if the hashes match, non-zero otherwise.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		expectedHashStr := args[1]
		log := logger.With("path", path, "command", "calc", "expected_hash", expectedHashStr)

		// Parse the expected hash from hex string
		expectedHash, err := hex.DecodeString(expectedHashStr)
		if err != nil {
			log.Error("Failed to parse expected hash", "error", err)
			// Write error to stderr so it's visible to users
			if _, writeErr := fmt.Fprintf(cmd.ErrOrStderr(), "Error: invalid hash format: %q (expected hexadecimal string)\n", expectedHashStr); writeErr != nil {
				log.Error("Failed to write error to stderr", "error", writeErr)
			}
			return fmt.Errorf("invalid hash format: %q (expected hexadecimal string): %w", expectedHashStr, err)
		}

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

		log.Info("Starting hash computation for verification")
		start := time.Now()

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
		computedHashStr := fmt.Sprintf("%x", result.Hash)
		log.Info("Hash computation completed",
			"duration", duration,
			"computed_hash", computedHashStr,
			"size", result.Size,
		)

		// Compare hashes
		if len(result.Hash) != len(expectedHash) {
			log.Error("Hash length mismatch",
				"computed_length", len(result.Hash),
				"expected_length", len(expectedHash),
			)
			writeErr := writeHashLengthMismatchOutput(cmd, len(result.Hash), len(expectedHash), computedHashStr, expectedHashStr)
			if writeErr != nil {
				log.Error("Failed to write hash length mismatch output", "error", writeErr)
			}
			return fmt.Errorf("hash length mismatch")
		}

		match := true
		for i := range result.Hash {
			if result.Hash[i] != expectedHash[i] {
				match = false
				break
			}
		}

		if match {
			log.Info("Hash verification successful", "hash", computedHashStr)
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Hash matches: %s\n", computedHashStr); err != nil {
				log.Error("Failed to write output to stdout", "error", err)
				return fmt.Errorf("failed to write output: %w", err)
			}
			return nil
		}

		log.Error("Hash verification failed",
			"computed_hash", computedHashStr,
			"expected_hash", expectedHashStr,
		)
		if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Hash mismatch!\n"); err != nil {
			log.Error("Failed to write output to stderr", "error", err)
			return fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Computed: %s\n", computedHashStr); err != nil {
			log.Error("Failed to write output to stderr", "error", err)
			return fmt.Errorf("failed to write output: %w", err)
		}
		if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Expected: %s\n", expectedHashStr); err != nil {
			log.Error("Failed to write output to stderr", "error", err)
			return fmt.Errorf("failed to write output: %w", err)
		}
		return fmt.Errorf("hash mismatch")
	},
}

// writeHashLengthMismatchOutput writes hash length mismatch information to stderr.
// It outputs the computed and expected hash lengths and values to help diagnose
// verification failures. This is a helper function to improve error handling consistency.
//
// Parameters:
//   - cmd: The Cobra command instance for accessing output streams
//   - computedLen: The length in bytes of the computed hash
//   - expectedLen: The length in bytes of the expected hash
//   - computedHash: The hexadecimal representation of the computed hash
//   - expectedHash: The hexadecimal representation of the expected hash
//
// Returns an error if writing to stderr fails.
func writeHashLengthMismatchOutput(cmd *cobra.Command, computedLen, expectedLen int, computedHash, expectedHash string) error {
	if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Hash mismatch: computed hash length (%d) differs from expected hash length (%d)\n",
		computedLen, expectedLen); err != nil {
		return fmt.Errorf("failed to write length mismatch: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Computed: %s\n", computedHash); err != nil {
		return fmt.Errorf("failed to write computed hash: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.OutOrStderr(), "Expected: %s\n", expectedHash); err != nil {
		return fmt.Errorf("failed to write expected hash: %w", err)
	}
	return nil
}

func init() {
	calcCmd.Flags().StringArrayP("exclude", "e", []string{}, "Exclude patterns (e.g., 'node_modules', '.git'). Can be specified multiple times.")
	calcCmd.Flags().StringP("ignore-file", "i", "", "Path to a custom ignore file (takes highest priority). .mtcignore and .gitignore are always loaded automatically from the working directory.")

	cmd.Register(calcCmd)
}

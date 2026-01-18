// Package merkle (diff.go) provides directory comparison functionality.
// It computes Merkle root hashes for two paths and compares them to detect differences.
package merkle

import (
	"bytes"
	"fmt"
	"time"

	"github.com/lucho00cuba/mtc/internal/logger"
)

const (
	// noDifferencesMsg is the message returned when two paths have identical hashes
	noDifferencesMsg = "No differences detected"
)

// Compare computes the Merkle root hashes of two paths and returns a list of differences.
// If the hashes are identical, it returns a message indicating no differences.
// Otherwise, it returns a message showing the hash mismatch.
// It automatically loads .mtcignore and .gitignore files from the working directory.
//
// This is a convenience function that uses default exclusion settings.
// For more control, use CompareWithExclusions.
//
// Parameters:
//   - a: The first path to compare (file or directory)
//   - b: The second path to compare (file or directory)
//
// Returns a slice of difference messages and any error encountered.
func Compare(a, b string) ([]string, error) {
	return CompareWithExclusions(a, b, nil, true, "")
}

// CompareWithExclusions computes the Merkle root hashes of two paths with exclusion patterns.
// It applies the same exclusion patterns to both paths to ensure fair comparison.
// The function computes hashes sequentially and compares the results.
//
// Parameters:
//   - a: The first path to compare (file or directory)
//   - b: The second path to compare (file or directory)
//   - patterns: Exclusion patterns to apply to both paths (e.g., "node_modules", ".git")
//   - loadIgnoreFile: If true, loads .mtcignore and .gitignore files from the working directory
//   - customIgnoreFile: Optional path to a custom ignore file (takes highest priority if provided)
//
// Returns a slice of difference messages. If paths are identical, returns a single
// "No differences detected" message. Otherwise, returns hash mismatch information.
func CompareWithExclusions(a, b string, patterns []string, loadIgnoreFile bool, customIgnoreFile string) ([]string, error) {
	log := logger.With("pathA", a, "pathB", b, "operation", "compare")

	// Create engines with exclusions for both paths
	var engineA, engineB *Engine
	var err error

	if len(patterns) > 0 || loadIgnoreFile || customIgnoreFile != "" {
		engineA, err = NewEngineWithExclusions(0, patterns, a, loadIgnoreFile, customIgnoreFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create engine for path A: %w", err)
		}
		engineB, err = NewEngineWithExclusions(0, patterns, b, loadIgnoreFile, customIgnoreFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create engine for path B: %w", err)
		}
	}

	log.Info("Starting hash computation for path A")
	startA := time.Now()
	var resultA Result
	if engineA != nil {
		resultA, err = engineA.HashPath(a)
	} else {
		resultA, err = HashPath(a)
	}
	if err != nil {
		log.Error("Failed to hash path A", "error", err, "duration", time.Since(startA))
		return nil, fmt.Errorf("failed to hash path %q: %w", a, err)
	}
	durationA := time.Since(startA)
	log.Info("Hash computation for path A completed",
		"duration", durationA,
		"hash", fmt.Sprintf("%x", resultA.Hash),
		"size", resultA.Size,
	)

	log.Info("Starting hash computation for path B")
	startB := time.Now()
	var resultB Result
	if engineB != nil {
		resultB, err = engineB.HashPath(b)
	} else {
		resultB, err = HashPath(b)
	}
	if err != nil {
		log.Error("Failed to hash path B", "error", err, "duration", time.Since(startB))
		return nil, fmt.Errorf("failed to hash path %q: %w", b, err)
	}
	durationB := time.Since(startB)
	log.Info("Hash computation for path B completed",
		"duration", durationB,
		"hash", fmt.Sprintf("%x", resultB.Hash),
		"size", resultB.Size,
	)

	if bytes.Equal(resultA.Hash, resultB.Hash) {
		log.Info("Paths are identical", "total_duration", durationA+durationB)
		return []string{noDifferencesMsg}, nil
	}

	log.Warn("Paths differ",
		"hashA", fmt.Sprintf("%x", resultA.Hash),
		"hashB", fmt.Sprintf("%x", resultB.Hash),
		"sizeA", resultA.Size,
		"sizeB", resultB.Size,
	)
	return []string{
		fmt.Sprintf("Root mismatch:\nA: %x (size: %d)\nB: %x (size: %d)",
			resultA.Hash, resultA.Size, resultB.Hash, resultB.Size),
	}, nil
}

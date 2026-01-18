// Package merkle provides Merkle tree hashing functionality for files and directories.
// It implements a deterministic hashing algorithm that builds a Merkle tree structure
// from directory contents, allowing for efficient integrity verification and comparison.
// The package uses BLAKE3 for hashing and supports exclusion patterns for filtering
// files and directories during hash computation.
package merkle

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lucho00cuba/mtc/internal/ignore"
	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/zeebo/blake3"
)

const (
	// DefaultBufferSize is the default buffer size for reading files
	DefaultBufferSize = 256 * 1024 // 256KB
	// DefaultMaxWorkers limits concurrent directory hashing to avoid IO thrashing
	DefaultMaxWorkers = 8
	// HashSize is the size in bytes of MTC node hashes.
	// BLAKE3 produces 32-byte (256-bit) hashes by default.
	HashSize = 32
)

// Result represents the result of hashing a path, containing both the hash and size.
// The hash is a BLAKE3 hash (32 bytes by default) representing the Merkle root,
// and the size is the total size in bytes of all files hashed.
type Result struct {
	// Hash is the Merkle root hash as a byte slice.
	// For files, this is the BLAKE3 hash of the file contents.
	// For directories, this is the combined hash of all entries.
	Hash []byte

	// Size is the total size in bytes of all files hashed.
	// For files, this is the file size.
	// For directories, this is the sum of all file sizes in the tree.
	Size int64
}

// Engine represents a Merkle hashing engine with configurable concurrency and buffer management.
// This structure is designed to be future-proof for caching, tree export, and partial diffing.
type Engine struct {
	maxWorkers int
	bufferPool *sync.Pool
	// sem is a global semaphore shared across the entire engine lifecycle.
	// It prevents goroutine/thread explosion by bounding concurrent hashing work.
	sem chan struct{}
	// matcher determines which paths should be excluded from hashing
	matcher ignore.Matcher
	// rootPath is the root path being hashed, used for computing relative paths for matching
	rootPath string
}

// NewEngine creates a new Merkle hashing engine with default settings.
func NewEngine() *Engine {
	return &Engine{
		maxWorkers: DefaultMaxWorkers,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, DefaultBufferSize)
				return &buf
			},
		},
		sem: make(chan struct{}, DefaultMaxWorkers),
	}
}

// NewEngineWithWorkers creates a new engine with a custom worker count.
func NewEngineWithWorkers(maxWorkers int) *Engine {
	if maxWorkers < 1 {
		maxWorkers = DefaultMaxWorkers
	}
	return &Engine{
		maxWorkers: maxWorkers,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, DefaultBufferSize)
				return &buf
			},
		},
		sem: make(chan struct{}, maxWorkers),
	}
}

// NewEngineWithExclusions creates a new engine with exclusion patterns.
// patterns are exclusion patterns (e.g., "node_modules", ".git").
// rootPath is the root path being hashed (used for computing relative paths and loading .mtcignore).
// loadIgnoreFile if true, loads .mtcignore and .gitignore files from the working directory.
// customIgnoreFile is an optional path to a custom ignore file (takes highest priority if provided).
func NewEngineWithExclusions(maxWorkers int, patterns []string, rootPath string, loadIgnoreFile bool, customIgnoreFile string) (*Engine, error) {
	matcher, err := ignore.NewMatcher(patterns, rootPath, loadIgnoreFile, customIgnoreFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create exclusion matcher: %w", err)
	}

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	if maxWorkers < 1 {
		maxWorkers = DefaultMaxWorkers
	}

	return &Engine{
		maxWorkers: maxWorkers,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, DefaultBufferSize)
				return &buf
			},
		},
		sem:      make(chan struct{}, maxWorkers),
		matcher:  matcher,
		rootPath: absRoot,
	}, nil
}

// HashPath computes the Merkle root hash and total size of a file or directory.
// For files, it returns the BLAKE3 hash of the file contents and its size.
// For directories, it recursively computes hashes of all entries and returns
// a combined hash representing the entire directory structure along with the total size.
// Symlinks are treated as leaf nodes; their target path is hashed, not traversed.
//
// This is a convenience function that creates a new engine with default settings.
// For more control over exclusions and concurrency, use Engine.HashPath directly.
//
// Parameters:
//   - path: The file or directory path to hash
//
// Returns the hash result and any error encountered during computation.
func HashPath(path string) (Result, error) {
	engine := NewEngine()
	return engine.HashPath(path)
}

// HashPath computes the Merkle root hash and total size using this engine instance.
// It sets the root path if not already set and uses the engine's configuration
// for exclusions and concurrency control.
//
// Parameters:
//   - path: The file or directory path to hash
//
// Returns the hash result and any error encountered during computation.
func (e *Engine) HashPath(path string) (Result, error) {
	// Set root path if not already set
	if e.rootPath == "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return Result{}, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		e.rootPath = absPath
	}

	visited := &sync.Map{}
	return e.hashPath(path, visited)
}

// hashPath is the internal implementation that tracks visited paths
// to prevent infinite loops with circular symlinks.
// It handles files, directories, and symlinks, applying exclusion patterns
// and building the Merkle tree structure recursively.
//
// Parameters:
//   - path: The file or directory path to hash (can be relative or absolute)
//   - visited: A thread-safe map tracking visited paths to detect circular symlinks
//
// Returns the hash result and any error encountered during computation.
func (e *Engine) hashPath(path string, visited *sync.Map) (Result, error) {
	// Resolve to absolute path to detect circular symlinks
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Result{}, fmt.Errorf("failed to resolve absolute path for %q: %w", path, err)
	}

	// Check for circular symlinks (thread-safe check)
	if _, exists := visited.Load(absPath); exists {
		logger.Error("Circular symlink detected", "path", absPath)
		return Result{}, fmt.Errorf("circular symlink detected at %q", absPath)
	}
	visited.Store(absPath, true)
	defer visited.Delete(absPath)

	info, err := os.Lstat(absPath)
	if err != nil {
		logger.Error("Failed to stat path", "path", absPath, "error", err)
		return Result{}, fmt.Errorf("failed to stat path %q: %w", absPath, err)
	}

	// Check if path should be excluded
	if e.matcher != nil {
		// Compute relative path from root for matching
		relPath, err := filepath.Rel(e.rootPath, absPath)
		if err != nil {
			// If we can't compute relative path, use the basename
			relPath = filepath.Base(absPath)
		}
		// Also check with absolute path and basename for flexibility
		if e.matcher.Match(relPath, info.IsDir()) ||
			e.matcher.Match(absPath, info.IsDir()) ||
			e.matcher.Match(filepath.Base(absPath), info.IsDir()) {
			logger.Debug("Excluding path", "path", absPath, "relative", relPath)
			// Return empty hash and zero size for excluded paths
			// This ensures excluded directories don't affect the hash
			h := blake3.New()
			return Result{Hash: h.Sum(nil), Size: 0}, nil
		}
	}

	// Treat symlinks as leaf nodes - hash their target path, don't traverse
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(absPath)
		if err != nil {
			logger.Error("Failed to read symlink", "path", absPath, "error", err)
			return Result{}, fmt.Errorf("failed to read symlink %q: %w", absPath, err)
		}
		// Hash the target path as a string (deterministic representation)
		h := blake3.New()
		if _, err := h.WriteString(target); err != nil {
			logger.Error("Failed to write to hash", "error", err)
			return Result{}, fmt.Errorf("failed to hash symlink target: %w", err)
		}
		logger.Debug("Hashed symlink as leaf node", "symlink", absPath, "target", target)
		// Symlinks have zero size
		return Result{Hash: h.Sum(nil), Size: 0}, nil
	}

	// After handling symlinks, check if it's a directory
	if info.IsDir() {
		logger.Debug("Processing directory", "path", absPath)
		return e.hashDir(absPath, visited)
	}

	logger.Debug("Processing file", "path", absPath, "size", info.Size())
	return e.hashFile(absPath, info.Size())
}

// hashFile computes the BLAKE3 hash of a file's contents using a pooled buffer.
// It validates the path is within the root directory to prevent directory traversal,
// acquires a semaphore slot to limit concurrent I/O, and uses a buffer pool for efficiency.
// It returns both the hash and the file size.
//
// Parameters:
//   - path: The absolute path to the file to hash
//   - size: The expected file size in bytes
//
// Returns the hash result and any error encountered during file reading or hashing.
func (e *Engine) hashFile(path string, size int64) (Result, error) {
	start := time.Now()
	log := logger.With("path", path, "operation", "hash_file")

	// Validate path is within rootPath to prevent directory traversal
	if e.rootPath != "" {
		cleanPath := filepath.Clean(path)
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return Result{}, fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		absRoot, err := filepath.Abs(e.rootPath)
		if err != nil {
			return Result{}, fmt.Errorf("failed to resolve root path: %w", err)
		}
		// Ensure the path is within the root directory
		relPath, err := filepath.Rel(absRoot, absPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return Result{}, fmt.Errorf("path outside allowed directory: %q", path)
		}
		path = absPath
	}

	// Acquire global semaphore to limit concurrent file hashing
	e.sem <- struct{}{}
	defer func() { <-e.sem }()

	f, err := os.Open(path)
	if err != nil {
		log.Error("Failed to open file", "error", err)
		return Result{}, fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warn("Failed to close file", "error", err)
		}
	}()

	// Get buffer from pool
	bufPtr, ok := e.bufferPool.Get().(*[]byte)
	if !ok {
		return Result{}, fmt.Errorf("failed to get buffer from pool")
	}
	defer e.bufferPool.Put(bufPtr)
	buf := *bufPtr

	h := blake3.New()
	bytesRead := int64(0)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			if _, writeErr := h.Write(buf[:n]); writeErr != nil {
				log.Error("Failed to write to hash", "error", writeErr)
				return Result{}, fmt.Errorf("failed to hash file content: %w", writeErr)
			}
			bytesRead += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Error("Failed to read file", "error", err, "bytes_read", bytesRead)
			return Result{}, fmt.Errorf("failed to read file %q: %w", path, err)
		}
	}

	duration := time.Since(start)
	log.Debug("File hashed successfully",
		"size", size,
		"bytes_read", bytesRead,
		"duration", duration,
	)

	return Result{Hash: h.Sum(nil), Size: size}, nil
}

// hashDir computes the Merkle root hash of a directory by hashing all entries
// in sorted order and combining their hashes. It also accumulates the total size.
// Entries are processed sequentially to maintain deterministic ordering.
// File hashing is bounded by a global semaphore to limit concurrent I/O.
//
// The function filters out special files (pipes, sockets, devices) and applies
// exclusion patterns before processing. Directory entries are sorted alphabetically
// to ensure deterministic hash computation.
//
// Parameters:
//   - path: The absolute path to the directory to hash
//   - visited: A thread-safe map tracking visited paths to detect circular symlinks
//
// Returns the hash result and any error encountered during directory processing.
func (e *Engine) hashDir(path string, visited *sync.Map) (Result, error) {
	start := time.Now()
	log := logger.With("path", path, "operation", "hash_dir")

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Error("Failed to read directory", "error", err)
		return Result{}, fmt.Errorf("failed to read directory %q: %w", path, err)
	}

	// Sort entries by name for deterministic hashing
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	log.Debug("Processing directory entries", "entry_count", len(entries))

	// Filter out special files and prepare work items
	type workItem struct {
		entry     os.DirEntry
		entryPath string
	}

	var workItems []workItem
	for _, entry := range entries {
		// Skip special files (pipes, sockets, devices) as they cannot be hashed
		if entry.Type()&(os.ModeNamedPipe|os.ModeSocket|os.ModeDevice) != 0 {
			log.Debug("Skipping special file", "entry", entry.Name(), "type", entry.Type())
			continue
		}

		childPath := filepath.Join(path, entry.Name())

		// Check if entry should be excluded
		if e.matcher != nil {
			relPath, err := filepath.Rel(e.rootPath, childPath)
			if err != nil {
				relPath = entry.Name()
			}
			isDir := entry.IsDir()
			if e.matcher.Match(relPath, isDir) ||
				e.matcher.Match(childPath, isDir) ||
				e.matcher.Match(entry.Name(), isDir) {
				log.Debug("Excluding entry", "entry", entry.Name(), "path", childPath)
				continue
			}
		}

		workItems = append(workItems, workItem{
			entry:     entry,
			entryPath: childPath,
		})
	}

	if len(workItems) == 0 {
		// Empty directory
		h := blake3.New()
		return Result{Hash: h.Sum(nil), Size: 0}, nil
	}

	// Sequentially process work items (no concurrency)
	results := make([]Result, len(workItems))

	for i, item := range workItems {
		entry := item.entry
		childPath := item.entryPath

		entryType := entry.Type()

		if entryType&os.ModeSymlink != 0 {
			target, err := os.Readlink(childPath)
			if err != nil {
				return Result{}, fmt.Errorf("failed to read symlink %q: %w", childPath, err)
			}
			h := blake3.New()
			if _, err := h.WriteString(target); err != nil {
				return Result{}, fmt.Errorf("failed to hash symlink target: %w", err)
			}
			results[i] = Result{Hash: h.Sum(nil), Size: 0}
			continue
		}

		if entry.IsDir() {
			result, err := e.hashPath(childPath, visited)
			if err != nil {
				return Result{}, fmt.Errorf("failed to hash entry %q in directory %q: %w", entry.Name(), path, err)
			}
			results[i] = result
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return Result{}, fmt.Errorf("failed to get info for entry %q in directory %q: %w", entry.Name(), path, err)
		}

		result, err := e.hashFile(childPath, info.Size())
		if err != nil {
			return Result{}, err
		}

		results[i] = result
	}

	// Combine all hashes and accumulate sizes
	h := blake3.New()
	var totalSize int64
	for _, result := range results {
		if _, err := h.Write(result.Hash); err != nil {
			log.Error("Failed to write to hash", "error", err)
			return Result{}, fmt.Errorf("failed to combine hashes: %w", err)
		}
		totalSize += result.Size
	}

	duration := time.Since(start)
	log.Debug("Directory hashed successfully",
		"entry_count", len(entries),
		"processed", len(workItems),
		"duration", duration,
		"total_size", totalSize,
	)

	return Result{Hash: h.Sum(nil), Size: totalSize}, nil
}

// Package ignore provides pattern matching functionality for excluding files and directories
// from hash computation. It supports .gitignore-style patterns including glob patterns,
// directory-only matches, and negation patterns. The package can load patterns from
// .mtcignore, .gitignore, and custom ignore files.
package ignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucho00cuba/mtc/internal/logger"
)

const (
	// globDoubleStar represents the "**" pattern that matches any number of directories
	globDoubleStar = "**"
)

// Matcher determines if a path should be excluded from hashing.
// Implementations of this interface provide pattern matching functionality
// to filter files and directories during hash computation.
type Matcher interface {
	// Match returns true if the path should be excluded from hashing.
	// The path can be relative to the root being hashed or absolute.
	//
	// Parameters:
	//   - path: The path to check (relative or absolute)
	//   - isDir: Whether the path represents a directory
	//
	// Returns true if the path matches an exclusion pattern and should be excluded.
	Match(path string, isDir bool) bool
}

// PatternMatcher matches paths against exclusion patterns.
// Supports patterns similar to .gitignore:
// - Exact matches: "node_modules"
// - Directory matches: "node_modules/" (matches directories only)
// - Glob patterns: "*.log", "**/build"
type PatternMatcher struct {
	patterns []pattern
}

type pattern struct {
	// raw is the original pattern string
	raw string
	// isDirOnly is true if pattern ends with /
	isDirOnly bool
	// isNegation is true if pattern starts with !
	isNegation bool
	// segments are the path segments to match
	segments []string
	// hasGlob is true if pattern contains * or ?
	hasGlob bool
}

// NewPatternMatcher creates a new pattern matcher from a list of patterns.
// Patterns support .gitignore-style syntax including:
//   - Exact matches: "node_modules"
//   - Directory-only: "node_modules/" (matches directories only)
//   - Glob patterns: "*.log", "**/build"
//   - Negation: "!important.log" (un-excludes previously excluded paths)
//
// Empty lines and lines starting with "#" are treated as comments and ignored.
//
// Parameters:
//   - patterns: A slice of pattern strings to compile
//
// Returns a new PatternMatcher instance ready to use.
func NewPatternMatcher(patterns []string) *PatternMatcher {
	pm := &PatternMatcher{
		patterns: make([]pattern, 0, len(patterns)),
	}

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" || strings.HasPrefix(p, "#") {
			continue // Skip empty lines and comments
		}

		pat := pattern{
			raw: p,
		}

		// Handle negation
		if strings.HasPrefix(p, "!") {
			pat.isNegation = true
			p = strings.TrimPrefix(p, "!")
		}

		// Handle directory-only patterns
		if strings.HasSuffix(p, "/") {
			pat.isDirOnly = true
			p = strings.TrimSuffix(p, "/")
		}

		// Normalize path separators
		p = filepath.ToSlash(p)
		pat.segments = strings.Split(p, "/")
		pat.hasGlob = strings.Contains(p, "*") || strings.Contains(p, "?")

		pm.patterns = append(pm.patterns, pat)
	}

	return pm
}

// Match returns true if the path should be excluded.
func (pm *PatternMatcher) Match(path string, isDir bool) bool {
	// Normalize path
	path = filepath.ToSlash(path)
	pathSegments := strings.Split(path, "/")

	// Track the most specific match (negation or exclusion)
	matched := false
	matchedNegation := false

	for _, pat := range pm.patterns {
		if pat.Match(pathSegments, isDir) {
			if pat.isNegation {
				matchedNegation = true
			} else {
				matched = true
			}
		}
	}

	// Negations override exclusions
	if matchedNegation {
		return false
	}
	return matched
}

// Match checks if the pattern matches the path segments.
func (p *pattern) Match(pathSegments []string, isDir bool) bool {
	// Directory-only patterns don't match files
	if p.isDirOnly && !isDir {
		return false
	}

	// Simple exact match for common cases
	if !p.hasGlob && len(p.segments) == 1 {
		// Check if any segment matches
		for _, seg := range pathSegments {
			if seg == p.segments[0] {
				return true
			}
		}
		return false
	}

	// For patterns with multiple segments or globs, use more complex matching
	return p.matchSegments(pathSegments)
}

// matchSegments performs pattern matching on path segments.
func (p *pattern) matchSegments(pathSegments []string) bool {
	patSegs := p.segments

	// Handle patterns starting with ** (match any number of directories)
	if len(patSegs) > 0 && patSegs[0] == globDoubleStar {
		// ** matches everything, so check if remaining pattern matches
		if len(patSegs) == 1 {
			return true
		}
		// Try matching remaining pattern at any position
		remainingPat := patSegs[1:]
		for i := 0; i <= len(pathSegments); i++ {
			if matchSegmentsAt(pathSegments[i:], remainingPat) {
				return true
			}
		}
		return false
	}

	// Handle patterns ending with **
	if len(patSegs) > 0 && patSegs[len(patSegs)-1] == globDoubleStar {
		// Match everything from the start
		return matchSegmentsAt(pathSegments, patSegs[:len(patSegs)-1])
	}

	// Standard matching from the end (most common case: "node_modules", ".git")
	// Check if pattern matches at the end of the path
	return matchSegmentsAt(pathSegments, patSegs)
}

// matchSegmentsAt checks if pattern segments match path segments starting at a given position.
func matchSegmentsAt(pathSegs []string, patSegs []string) bool {
	if len(patSegs) == 0 {
		return true
	}
	if len(pathSegs) == 0 {
		return false
	}

	// Try matching pattern at any position in the path
	// This handles cases like "node_modules" appearing anywhere in the path
	for i := 0; i <= len(pathSegs)-len(patSegs); i++ {
		matched := true
		for j := 0; j < len(patSegs); j++ {
			if !matchSegment(pathSegs[i+j], patSegs[j]) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}

	return false
}

// matchSegment checks if a single path segment matches a pattern segment.
func matchSegment(pathSeg, patSeg string) bool {
	// Exact match
	if patSeg == pathSeg {
		return true
	}

	// Simple glob matching
	if strings.Contains(patSeg, "*") || strings.Contains(patSeg, "?") {
		return matchGlob(pathSeg, patSeg)
	}

	return false
}

// matchGlob performs simple glob matching.
func matchGlob(s, pattern string) bool {
	// Convert pattern to regex-like matching
	// * matches any sequence, ? matches any single character
	patternIdx := 0
	strIdx := 0

	for patternIdx < len(pattern) && strIdx < len(s) {
		if pattern[patternIdx] == '*' {
			// * matches everything, try matching rest of pattern
			if patternIdx == len(pattern)-1 {
				return true
			}
			// Try matching remaining pattern at each position
			for i := strIdx; i <= len(s); i++ {
				if matchGlob(s[i:], pattern[patternIdx+1:]) {
					return true
				}
			}
			return false
		} else if pattern[patternIdx] == '?' {
			// ? matches any single character
			patternIdx++
			strIdx++
		} else if pattern[patternIdx] == s[strIdx] {
			patternIdx++
			strIdx++
		} else {
			return false
		}
	}

	// Handle trailing *
	for patternIdx < len(pattern) && pattern[patternIdx] == '*' {
		patternIdx++
	}

	return patternIdx == len(pattern) && strIdx == len(s)
}

// LoadIgnoreFile loads patterns from an ignore file (.mtcignore or .gitignore).
// The function validates the filename to prevent directory traversal attacks
// and ensures the file is within the root directory. If the file doesn't exist,
// it returns nil without an error (treating it as no patterns).
//
// Parameters:
//   - rootPath: The root directory path where the ignore file should be located
//   - filename: The name of the ignore file (e.g., ".mtcignore", ".gitignore")
//
// Returns a slice of pattern strings and any error encountered.
// Returns nil, nil if the file doesn't exist (not an error condition).
func LoadIgnoreFile(rootPath string, filename string) ([]string, error) {
	// Clean and validate paths to prevent directory traversal
	cleanRoot := filepath.Clean(rootPath)
	cleanFilename := filepath.Clean(filename)

	// Ensure filename doesn't contain path separators or directory traversal (only allow simple filenames)
	if strings.Contains(filename, "..") || strings.Contains(filename, string(filepath.Separator)) || cleanFilename != filename {
		return nil, fmt.Errorf("invalid filename: %s", filename)
	}

	ignorePath := filepath.Join(cleanRoot, cleanFilename)

	// Resolve to absolute path and validate it's within rootPath
	absIgnorePath, err := filepath.Abs(ignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	absRoot, err := filepath.Abs(cleanRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	// Normalize paths by cleaning them to ensure consistent comparison
	absIgnorePath = filepath.Clean(absIgnorePath)
	absRoot = filepath.Clean(absRoot)

	// Ensure the ignore file path is within the root directory
	// The file should have the root path as a prefix followed by a separator
	// or be exactly at the root (which is impossible for a file, but we check anyway)
	// Special case: when root is "/", we need to handle it differently to avoid "//"
	var rootWithSep string
	if absRoot == string(filepath.Separator) || absRoot == "/" {
		// Root is filesystem root, so the file path should start with "/"
		rootWithSep = string(filepath.Separator)
	} else {
		rootWithSep = absRoot + string(filepath.Separator)
	}

	if absIgnorePath != absRoot && !strings.HasPrefix(absIgnorePath, rootWithSep) {
		return nil, fmt.Errorf("ignore file path outside root directory: %s", filename)
	}

	// absIgnorePath is validated to be within absRoot, safe to open
	//nolint:gosec // Path is validated to be within root directory above
	file, err := os.Open(absIgnorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist, no patterns
		}
		return nil, fmt.Errorf("failed to open %s: %w", filename, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Warn("Failed to close ignore file", "error", err)
		}
	}()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	logger.Info("Loaded ignore file", "file", ignorePath, "patterns", len(patterns), "filename", filename)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}

	return patterns, nil
}

// FindIgnoreFiles searches for .mtcignore and .gitignore files from the working directory up to the root.
// It walks up the directory tree starting from the current working directory
// (where the command is executed), not from the path being hashed.
//
// Returns patterns from all found ignore files. Patterns from directories closer
// to the root take precedence. .mtcignore patterns take precedence over .gitignore patterns.
//
// Returns a slice of all collected patterns and any error encountered during the search.
func FindIgnoreFiles() ([]string, error) {
	var allPatterns []string

	// Get current working directory (where the command is executed from)
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	absPath, err := filepath.Abs(wd)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Start from the working directory and walk up to root
	current := absPath
	visited := make(map[string]bool)

	for {
		// Check if we've already processed this directory
		if visited[current] {
			break
		}
		visited[current] = true

		// Try to load .mtcignore first (has priority)
		mtcPatterns, err := LoadIgnoreFile(current, ".mtcignore")
		if err != nil {
			return nil, err
		}
		if mtcPatterns != nil {
			// Prepend patterns from closer directories (they take precedence)
			allPatterns = append(mtcPatterns, allPatterns...)
		}

		// Try to load .gitignore (only if .mtcignore doesn't exist or as supplement)
		gitPatterns, err := LoadIgnoreFile(current, ".gitignore")
		if err != nil {
			return nil, err
		}
		if gitPatterns != nil {
			// Append .gitignore patterns after .mtcignore (lower priority)
			allPatterns = append(allPatterns, gitPatterns...)
		}

		// Move to parent directory
		parent := filepath.Dir(current)
		if parent == current {
			break // Reached filesystem root
		}
		current = parent
	}

	return allPatterns, nil
}

// LoadCustomIgnoreFile loads patterns from a custom ignore file specified by the user.
// The file path is validated and normalized to prevent directory traversal attacks.
// Unlike LoadIgnoreFile, this function returns an error if the file doesn't exist,
// as the user explicitly specified the file path.
//
// Parameters:
//   - filePath: The absolute or relative path to the custom ignore file
//
// Returns a slice of pattern strings and any error encountered.
// Returns an error if the file doesn't exist or cannot be read.
func LoadCustomIgnoreFile(filePath string) ([]string, error) {
	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(filePath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Validate that the cleaned absolute path doesn't contain directory traversal
	// After filepath.Clean and filepath.Abs, the path should be normalized
	// Additional check: ensure the resolved path matches the cleaned path
	if absPath != filepath.Clean(absPath) {
		return nil, fmt.Errorf("invalid file path: %s", filePath)
	}

	// Validate that the path doesn't attempt to escape (double-check after normalization)
	// This is a user-provided path, so we validate it's a legitimate file path
	if strings.Contains(absPath, "..") {
		return nil, fmt.Errorf("invalid file path: %s", filePath)
	}

	// absPath is validated and normalized, safe to open
	// Path is validated and normalized above
	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("ignore file does not exist: %s", filePath)
		}
		return nil, fmt.Errorf("failed to open ignore file %s: %w", filePath, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Warn("Failed to close ignore file", "error", err)
		}
	}()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read ignore file %s: %w", filePath, err)
	}

	return patterns, nil
}

// NewMatcher creates a matcher from patterns and optionally loads .mtcignore and .gitignore files.
// It combines patterns from multiple sources in the following priority order (highest to lowest):
//  1. Custom ignore file (if provided)
//  2. Command-line exclusion patterns
//  3. .mtcignore and .gitignore files (if loadIgnoreFile is true)
//
// Ignore files are loaded from the current working directory (where the command is executed),
// not from the rootPath being hashed. This allows ignore files to be placed in the project root
// regardless of which directory is being hashed.
//
// Parameters:
//   - patterns: Command-line exclusion patterns to include
//   - rootPath: The root path being hashed (used for context, not for loading ignore files)
//   - loadIgnoreFile: If true, automatically loads .mtcignore and .gitignore files
//   - customIgnoreFile: Optional path to a custom ignore file (always loaded if provided)
//
// Returns a Matcher instance ready to use, or an error if pattern compilation fails.
func NewMatcher(patterns []string, rootPath string, loadIgnoreFile bool, customIgnoreFile string) (Matcher, error) {
	allPatterns := make([]string, len(patterns))
	copy(allPatterns, patterns)

	var customPatterns []string
	var ignorePatterns []string

	// Load custom ignore file first (highest priority, always loaded if specified)
	if customIgnoreFile != "" {
		var err error
		customPatterns, err = LoadCustomIgnoreFile(customIgnoreFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom ignore file: %w", err)
		}
		allPatterns = append(allPatterns, customPatterns...)
		logger.Info("Loaded custom ignore file", "file", customIgnoreFile, "patterns", len(customPatterns))
	}

	// Load automatic ignore files (.mtcignore and .gitignore) only if loadIgnoreFile is true
	if loadIgnoreFile {
		var err error
		ignorePatterns, err = FindIgnoreFiles()
		if err != nil {
			return nil, fmt.Errorf("failed to load ignore files: %w", err)
		}
		allPatterns = append(allPatterns, ignorePatterns...)
		if len(ignorePatterns) > 0 {
			logger.Info("Loaded automatic ignore files", "patterns", len(ignorePatterns))
		}
	}

	if len(allPatterns) == 0 {
		return &noOpMatcher{}, nil
	}

	return NewPatternMatcher(allPatterns), nil
}

// noOpMatcher is a Matcher implementation that never matches anything.
// It is used when no exclusion patterns are provided, allowing all paths
// to be included in hash computation.
type noOpMatcher struct{}

// Match always returns false, indicating no paths should be excluded.
// This allows all files and directories to be processed when no exclusions are configured.
//
// Parameters:
//   - path: The path to check (unused)
//   - isDir: Whether the path is a directory (unused)
//
// Returns false (never excludes anything).
func (n *noOpMatcher) Match(path string, isDir bool) bool {
	return false
}

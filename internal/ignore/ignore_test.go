package ignore

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucho00cuba/mtc/internal/logger"
)

func init() {
	// Silence logger during tests - only show errors
	logger.Init("error", "text", io.Discard)
}

func TestNewPatternMatcher(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     int // expected number of patterns after filtering
	}{
		{
			name:     "empty patterns",
			patterns: []string{},
			want:     0,
		},
		{
			name:     "single pattern",
			patterns: []string{"node_modules"},
			want:     1,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"node_modules", ".git", "dist"},
			want:     3,
		},
		{
			name:     "with comments",
			patterns: []string{"# comment", "node_modules", "# another comment"},
			want:     1,
		},
		{
			name:     "with empty lines",
			patterns: []string{"", "node_modules", "  ", ".git"},
			want:     2,
		},
		{
			name:     "with negation",
			patterns: []string{"!important", "*.log"},
			want:     2,
		},
		{
			name:     "with directory pattern",
			patterns: []string{"node_modules/", "*.log"},
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPatternMatcher(tt.patterns)
			if len(pm.patterns) != tt.want {
				t.Errorf("NewPatternMatcher() got %d patterns, want %d", len(pm.patterns), tt.want)
			}
		})
	}
}

func TestPatternMatcher_Match(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		isDir    bool
		want     bool
	}{
		// Exact matches
		{
			name:     "exact match file",
			patterns: []string{"test.txt"},
			path:     "test.txt",
			isDir:    false,
			want:     true,
		},
		{
			name:     "exact match in path",
			patterns: []string{"node_modules"},
			path:     "project/node_modules/package",
			isDir:    false,
			want:     true,
		},
		{
			name:     "no match",
			patterns: []string{"node_modules"},
			path:     "project/src/main.go",
			isDir:    false,
			want:     false,
		},
		// Directory-only patterns
		{
			name:     "directory pattern matches dir",
			patterns: []string{"node_modules/"},
			path:     "project/node_modules",
			isDir:    true,
			want:     true,
		},
		{
			name:     "directory pattern doesn't match file",
			patterns: []string{"node_modules/"},
			path:     "project/node_modules",
			isDir:    false,
			want:     false,
		},
		// Glob patterns
		{
			name:     "glob match *.log",
			patterns: []string{"*.log"},
			path:     "app.log",
			isDir:    false,
			want:     true,
		},
		{
			name:     "glob match in path",
			patterns: []string{"*.log"},
			path:     "logs/app.log",
			isDir:    false,
			want:     true,
		},
		{
			name:     "glob no match",
			patterns: []string{"*.log"},
			path:     "app.txt",
			isDir:    false,
			want:     false,
		},
		{
			name:     "glob with ?",
			patterns: []string{"test?.txt"},
			path:     "test1.txt",
			isDir:    false,
			want:     true,
		},
		// Negation
		{
			name:     "negation overrides exclusion",
			patterns: []string{"*.log", "!important.log"},
			path:     "important.log",
			isDir:    false,
			want:     false,
		},
		{
			name:     "negation doesn't affect other files",
			patterns: []string{"*.log", "!important.log"},
			path:     "other.log",
			isDir:    false,
			want:     true,
		},
		// Multiple patterns
		{
			name:     "multiple patterns match",
			patterns: []string{"node_modules", ".git"},
			path:     ".git",
			isDir:    true,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPatternMatcher(tt.patterns)
			got := pm.Match(tt.path, tt.isDir)
			if got != tt.want {
				t.Errorf("PatternMatcher.Match(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestLoadIgnoreFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		content   string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid ignore file",
			filename:  ".mtcignore",
			content:   "node_modules\n.git\n*.log\n",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "file with comments",
			filename:  ".mtcignore",
			content:   "# comment\nnode_modules\n# another\n.git\n",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "file with empty lines",
			filename:  ".mtcignore",
			content:   "node_modules\n\n.git\n  \n",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			filename:  ".nonexistent",
			content:   "",
			wantCount: 0,
			wantErr:   false, // Should return nil, nil for non-existent files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content != "" {
				filePath := filepath.Join(tmpDir, tt.filename)
				err := os.WriteFile(filePath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			patterns, err := LoadIgnoreFile(tmpDir, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadIgnoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(patterns) != tt.wantCount {
				t.Errorf("LoadIgnoreFile() got %d patterns, want %d", len(patterns), tt.wantCount)
			}
		})
	}
}

func TestLoadCustomIgnoreFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantCount int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid file",
			content:   "node_modules\n.git\n",
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "non-existent file",
			content:   "",
			wantCount: 0,
			wantErr:   true,
			errMsg:    "ignore file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.content != "" {
				filePath = filepath.Join(tmpDir, "custom.ignore")
				err := os.WriteFile(filePath, []byte(tt.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			} else {
				filePath = filepath.Join(tmpDir, "nonexistent.ignore")
			}

			patterns, err := LoadCustomIgnoreFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadCustomIgnoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadCustomIgnoreFile() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if len(patterns) != tt.wantCount {
					t.Errorf("LoadCustomIgnoreFile() got %d patterns, want %d", len(patterns), tt.wantCount)
				}
			}
		})
	}
}

func TestNewMatcher(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test ignore files
	mtcContent := "node_modules\n.git\n"
	mtcPath := filepath.Join(tmpDir, ".mtcignore")
	if err := os.WriteFile(mtcPath, []byte(mtcContent), 0644); err != nil {
		t.Fatalf("Failed to create .mtcignore: %v", err)
	}

	gitContent := "*.log\n*.tmp\n"
	gitPath := filepath.Join(tmpDir, ".gitignore")
	if err := os.WriteFile(gitPath, []byte(gitContent), 0644); err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Change to temp directory to test FindIgnoreFiles
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	tests := []struct {
		name             string
		patterns         []string
		loadIgnoreFile   bool
		customIgnoreFile string
		wantErr          bool
	}{
		{
			name:             "with patterns only",
			patterns:         []string{"test"},
			loadIgnoreFile:   false,
			customIgnoreFile: "",
			wantErr:          false,
		},
		{
			name:             "with loadIgnoreFile",
			patterns:         []string{},
			loadIgnoreFile:   true,
			customIgnoreFile: "",
			wantErr:          false,
		},
		{
			name:             "with custom ignore file",
			patterns:         []string{},
			loadIgnoreFile:   false,
			customIgnoreFile: mtcPath,
			wantErr:          false,
		},
		{
			name:             "empty matcher",
			patterns:         []string{},
			loadIgnoreFile:   false,
			customIgnoreFile: "",
			wantErr:          false,
		},
		{
			name:             "invalid custom file",
			patterns:         []string{},
			loadIgnoreFile:   false,
			customIgnoreFile: filepath.Join(tmpDir, "nonexistent"),
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, err := NewMatcher(tt.patterns, tmpDir, tt.loadIgnoreFile, tt.customIgnoreFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && matcher == nil {
				t.Error("NewMatcher() returned nil matcher without error")
			}
		})
	}
}

func TestNoOpMatcher(t *testing.T) {
	matcher := &noOpMatcher{}

	if matcher.Match("anything", true) {
		t.Error("noOpMatcher.Match() should always return false")
	}
	if matcher.Match("anything", false) {
		t.Error("noOpMatcher.Match() should always return false")
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		str     string
		want    bool
	}{
		{
			name:    "simple wildcard",
			pattern: "*.log",
			str:     "app.log",
			want:    true,
		},
		{
			name:    "wildcard no match",
			pattern: "*.log",
			str:     "app.txt",
			want:    false,
		},
		{
			name:    "question mark match",
			pattern: "test?.txt",
			str:     "test1.txt",
			want:    true,
		},
		{
			name:    "question mark no match",
			pattern: "test?.txt",
			str:     "test12.txt",
			want:    false,
		},
		{
			name:    "multiple wildcards",
			pattern: "*.*",
			str:     "file.txt",
			want:    true,
		},
		{
			name:    "trailing wildcard",
			pattern: "prefix*",
			str:     "prefix123",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGlob(tt.str, tt.pattern)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.str, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestPatternMatchSegments(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		pathSegments []string
		isDir        bool
		want         bool
	}{
		{
			name:         "simple match",
			pattern:      "node_modules",
			pathSegments: []string{"project", "node_modules"},
			isDir:        true,
			want:         true,
		},
		{
			name:         "match with **",
			pattern:      "**/build",
			pathSegments: []string{"project", "src", "build"},
			isDir:        true,
			want:         true,
		},
		{
			name:         "match ending with **",
			pattern:      "src/**",
			pathSegments: []string{"project", "src", "file.go"},
			isDir:        false,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewPatternMatcher([]string{tt.pattern})
			if len(pm.patterns) == 0 {
				t.Fatal("Pattern not created")
			}
			pat := pm.patterns[0]
			got := pat.matchSegments(tt.pathSegments)
			if got != tt.want {
				t.Errorf("pattern.matchSegments() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

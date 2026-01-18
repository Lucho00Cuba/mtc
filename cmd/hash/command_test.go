package hash

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucho00cuba/mtc/cmd"
	"github.com/lucho00cuba/mtc/internal/logger"
)

func init() {
	// Silence logger during tests - only show errors
	logger.Init("error", "text", io.Discard)
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0 B",
		},
		{
			name:  "less than 1KB",
			bytes: 512,
			want:  "512 B",
		},
		{
			name:  "exactly 1KB",
			bytes: 1024,
			want:  "1 KB",
		},
		{
			name:  "1.5KB",
			bytes: 1536,
			want:  "1.5 KB",
		},
		{
			name:  "1MB",
			bytes: 1024 * 1024,
			want:  "1.0 MB",
		},
		{
			name:  "1.5MB",
			bytes: 1024 * 1024 * 1.5,
			want:  "1.5 MB",
		},
		{
			name:  "1GB",
			bytes: 1024 * 1024 * 1024,
			want:  "1.0 GB",
		},
		{
			name:  "large size",
			bytes: 1024 * 1024 * 1024 * 5,
			want:  "5.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestHashCmd_File(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"hash", testFile})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, testFile) {
		t.Errorf("Output should contain file path, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
	if !strings.Contains(output, "(f):") {
		t.Errorf("Output should indicate file type, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestHashCmd_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"hash", tmpDir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, tmpDir) {
		t.Errorf("Output should contain directory path, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
	if !strings.Contains(output, "(d):") {
		t.Errorf("Output should indicate directory type, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestHashCmd_Nonexistent(t *testing.T) {
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetArgs([]string{"hash", "/nonexistent/path/that/does/not/exist"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("rootCmd.Execute() expected error for nonexistent path")
	}
}

func TestHashCmd_WithExcludeFlag(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatalf("Failed to create keep.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "exclude.txt"), []byte("exclude"), 0644); err != nil {
		t.Fatalf("Failed to create exclude.txt: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"hash", "-e", "exclude.txt", tmpDir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with exclude flag error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, tmpDir) {
		t.Errorf("Output should contain directory path, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestHashCmd_WithIgnoreFileFlag(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	ignoreFile := filepath.Join(tmpDir, "custom.ignore")
	if err := os.WriteFile(ignoreFile, []byte("*.txt\n"), 0644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"hash", "-i", ignoreFile, tmpDir})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with ignore file flag error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, tmpDir) {
		t.Errorf("Output should contain directory path, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestHashCmd_InvalidArgs(t *testing.T) {
	// Verify that Args validator is set
	if hashCmd.Args == nil {
		t.Fatal("hashCmd should have Args validator set")
	}

	// Test with no args - should return error
	err := hashCmd.Args(hashCmd, []string{})
	if err == nil {
		t.Error("hashCmd.Args() expected error for no args")
	}

	// Test with too many args - should return error
	err = hashCmd.Args(hashCmd, []string{"arg1", "arg2"})
	if err == nil {
		t.Error("hashCmd.Args() expected error for too many args")
	}

	// Test with correct number of args - should not error
	err = hashCmd.Args(hashCmd, []string{"path"})
	if err != nil {
		t.Errorf("hashCmd.Args() unexpected error for valid args: %v", err)
	}
}

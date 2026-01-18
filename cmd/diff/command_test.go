package diff

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

func TestDiffCmd_Identical(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create identical files
	if err := os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("same content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("same content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"diff", dir1, dir2})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	output := buf.String()
	// Also check stderr in case output went there
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, "No differences") {
		t.Errorf("Output should indicate no differences, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestDiffCmd_Different(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create different files
	if err := os.WriteFile(filepath.Join(dir1, "file.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "file.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"diff", dir1, dir2})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if strings.Contains(output, "No differences") {
		t.Errorf("Output should indicate differences, got: %s", output)
	}
	if !strings.Contains(output, "Root mismatch") {
		t.Errorf("Output should contain mismatch message, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestDiffCmd_Nonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	nonexistent := filepath.Join(tmpDir, "nonexistent")

	rootCmd := cmd.GetRootCmd()
	rootCmd.SetArgs([]string{"diff", nonexistent, tmpDir})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("rootCmd.Execute() expected error for nonexistent path")
	}
}

func TestDiffCmd_WithExcludeFlag(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	// Create same files
	if err := os.WriteFile(filepath.Join(dir1, "keep.txt"), []byte("same"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "keep.txt"), []byte("same"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Create different excluded files
	if err := os.WriteFile(filepath.Join(dir1, "exclude.txt"), []byte("different1"), 0644); err != nil {
		t.Fatalf("Failed to create exclude file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "exclude.txt"), []byte("different2"), 0644); err != nil {
		t.Fatalf("Failed to create exclude file: %v", err)
	}

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"diff", "-e", "exclude.txt", dir1, dir2})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with exclude flag error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if !strings.Contains(output, "No differences") {
		t.Errorf("Output should indicate no differences when excluded files differ, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestDiffCmd_WithIgnoreFileFlag(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatalf("Failed to create dir1: %v", err)
	}
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatalf("Failed to create dir2: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir1, "test.txt"), []byte("same"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir2, "test.txt"), []byte("same"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
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
	rootCmd.SetArgs([]string{"diff", "-i", ignoreFile, dir1, dir2})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with ignore file flag error = %v", err)
	}

	output := buf.String()
	if errBuf.Len() > 0 {
		output = errBuf.String() + output
	}
	if output == "" {
		t.Errorf("Output should not be empty, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestDiffCmd_InvalidArgs(t *testing.T) {
	// Verify that Args validator is set
	if diffCmd.Args == nil {
		t.Fatal("diffCmd should have Args validator set")
	}

	// Test with no args - should return error
	err := diffCmd.Args(diffCmd, []string{})
	if err == nil {
		t.Error("diffCmd.Args() expected error for no args")
	}

	// Test with one arg - should return error
	err = diffCmd.Args(diffCmd, []string{"arg1"})
	if err == nil {
		t.Error("diffCmd.Args() expected error for one arg")
	}

	// Test with too many args - should return error
	err = diffCmd.Args(diffCmd, []string{"arg1", "arg2", "arg3"})
	if err == nil {
		t.Error("diffCmd.Args() expected error for too many args")
	}

	// Test with correct number of args - should not error
	err = diffCmd.Args(diffCmd, []string{"path1", "path2"})
	if err != nil {
		t.Errorf("diffCmd.Args() unexpected error for valid args: %v", err)
	}
}

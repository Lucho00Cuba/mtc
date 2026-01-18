package calc

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucho00cuba/mtc/cmd"
	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/lucho00cuba/mtc/internal/merkle"
)

func init() {
	// Silence logger during tests - only show errors
	logger.Init("error", "text", io.Discard)
}

func TestCalcCmd_MatchingHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Compute the expected hash
	engine, err := merkle.NewEngineWithExclusions(0, []string{}, testFile, true, "")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	result, err := engine.HashPath(testFile)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}
	expectedHash := hex.EncodeToString(result.Hash)

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"calc", testFile, expectedHash})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v, stderr: %s", err, errBuf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Hash matches:") {
		t.Errorf("Output should indicate hash match, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
	if !strings.Contains(output, expectedHash) {
		t.Errorf("Output should contain the hash, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestCalcCmd_MismatchingHash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Use a wrong hash
	wrongHash := "0000000000000000000000000000000000000000000000000000000000000000"

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"calc", testFile, wrongHash})

	err := rootCmd.Execute()
	// The command should exit with non-zero code, so we expect an error
	if err == nil {
		t.Error("rootCmd.Execute() expected error for mismatching hash")
	}

	// Check both stdout and stderr as cobra may redirect output
	output := buf.String() + errBuf.String()
	if !strings.Contains(output, "Hash mismatch!") {
		t.Errorf("Output should indicate hash mismatch, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestCalcCmd_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Compute the expected hash
	engine, err := merkle.NewEngineWithExclusions(0, []string{}, tmpDir, true, "")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	result, err := engine.HashPath(tmpDir)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}
	expectedHash := hex.EncodeToString(result.Hash)

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"calc", tmpDir, expectedHash})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() error = %v, stderr: %s", err, errBuf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Hash matches:") {
		t.Errorf("Output should indicate hash match, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestCalcCmd_InvalidHashFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Use an invalid hash format (not hex)
	invalidHash := "not-a-valid-hex-string"

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"calc", testFile, invalidHash})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("rootCmd.Execute() expected error for invalid hash format")
	}

	output := errBuf.String()
	if !strings.Contains(output, "invalid hash format") {
		t.Errorf("Output should indicate invalid hash format, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

func TestCalcCmd_NonexistentPath(t *testing.T) {
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetArgs([]string{"calc", "/nonexistent/path/that/does/not/exist", "0000000000000000000000000000000000000000000000000000000000000000"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("rootCmd.Execute() expected error for nonexistent path")
	}
}

func TestCalcCmd_InvalidArgs(t *testing.T) {
	// Verify that Args validator is set
	if calcCmd.Args == nil {
		t.Fatal("calcCmd should have Args validator set")
	}

	// Test with no args - should return error
	err := calcCmd.Args(calcCmd, []string{})
	if err == nil {
		t.Error("calcCmd.Args() expected error for no args")
	}

	// Test with one arg - should return error
	err = calcCmd.Args(calcCmd, []string{"arg1"})
	if err == nil {
		t.Error("calcCmd.Args() expected error for one arg")
	}

	// Test with too many args - should return error
	err = calcCmd.Args(calcCmd, []string{"arg1", "arg2", "arg3"})
	if err == nil {
		t.Error("calcCmd.Args() expected error for too many args")
	}

	// Test with correct number of args - should not error
	err = calcCmd.Args(calcCmd, []string{"path", "hash"})
	if err != nil {
		t.Errorf("calcCmd.Args() unexpected error for valid args: %v", err)
	}
}

func TestCalcCmd_WithExcludeFlag(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0644); err != nil {
		t.Fatalf("Failed to create keep.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "exclude.txt"), []byte("exclude"), 0644); err != nil {
		t.Fatalf("Failed to create exclude.txt: %v", err)
	}

	// Compute the expected hash with exclusions
	engine, err := merkle.NewEngineWithExclusions(0, []string{"exclude.txt"}, tmpDir, true, "")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	result, err := engine.HashPath(tmpDir)
	if err != nil {
		t.Fatalf("Failed to compute hash: %v", err)
	}
	expectedHash := hex.EncodeToString(result.Hash)

	var buf bytes.Buffer
	var errBuf bytes.Buffer
	rootCmd := cmd.GetRootCmd()
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs([]string{"calc", "-e", "exclude.txt", tmpDir, expectedHash})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with exclude flag error = %v, stderr: %s", err, errBuf.String())
	}

	output := buf.String()
	if !strings.Contains(output, "Hash matches:") {
		t.Errorf("Output should indicate hash match, got stdout: %q, stderr: %q", buf.String(), errBuf.String())
	}
}

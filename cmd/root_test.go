package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/spf13/cobra"
)

func init() {
	// Silence logger during tests - only show errors
	logger.Init("error", "text", io.Discard)
}

func TestRegister(t *testing.T) {
	// Create a test command
	testCmd := &cobra.Command{
		Use: "test",
	}

	// Register it
	Register(testCmd)

	// Verify it was added
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "test" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Register() should add command to rootCmd")
	}
}

func TestRootCmd_Help(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with --help error = %v", err)
	}

	output := buf.String()
	if !contains(output, "MTC") {
		t.Errorf("Help output should contain 'MTC', got: %s", output)
	}
}

func TestRootCmd_Version(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("rootCmd.Execute() with --version error = %v", err)
	}

	output := buf.String()
	if !contains(output, "mtc") {
		t.Errorf("Version output should contain 'mtc', got: %s", output)
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

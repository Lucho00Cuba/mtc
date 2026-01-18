// Package cmd provides the root command and command registration functionality
// for the MTC CLI application. It handles global flags, logging configuration,
// and command initialization.
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lucho00cuba/mtc/internal/logger"
	"github.com/lucho00cuba/mtc/version"
	"github.com/spf13/cobra"
)

var (
	// logLevel stores the logging level flag value.
	logLevel string

	// logFormat stores the logging format flag value (text or json).
	logFormat string

	// logOutput stores the log output destination flag value (stdout or filename).
	logOutput string

	// verbose stores the count of -v flags (0, 1, or 2).
	verbose int

	// quiet stores the quiet mode flag value.
	quiet bool

	// logFile stores the opened log file handle when logging to a file.
	logFile *os.File
)

// rootCmd is the root command for the mtc CLI application.
// It provides the main entry point and handles global configuration.
var rootCmd = &cobra.Command{
	Use:   "mtc",
	Short: "MTC (Merkle Tree Checksum) - generates deterministic directory checksums using Merkle Trees",
	Long: `MTC (Merkle Tree Checksum) is a deterministic directory checksum tool that generates
directory checksums using Merkle Trees. It provides a set of commands to interact with directories.`,
	Example: `  # Generate the checksum for the current directory
  mtc hash .

  # Generate the checksum for a directory, excluding node_modules and .git
  mtc hash /my/project -e node_modules -e .git

  # Compare two directories
  mtc diff /my/project /my/project-copy

  # Use a custom ignore file
  mtc hash . --ignore-file=myignore.txt

  # Verify a directory matches an expected hash
  mtc calc /my/project abc123def456...`,
	Version: version.VERSION,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Determine log level based on flags
		level := logLevel
		if quiet {
			level = "error"
		} else if verbose > 0 {
			// -v = info, -vv = debug
			if verbose >= 2 {
				level = "debug"
			} else {
				level = "info"
			}
		} else if level == "" {
			// Default to warn level when no verbose flag is set
			// This means Info and Debug logs won't be shown unless -v or -vv is used
			level = "warn"
		}

		// Determine log output destination
		var output io.Writer
		if logOutput == "" || logOutput == "stdout" {
			output = os.Stdout
		} else {
			// Clean and validate log file path to prevent directory traversal
			cleanPath := filepath.Clean(logOutput)
			absPath, err := filepath.Abs(cleanPath)
			if err != nil {
				return fmt.Errorf("error resolving log file path %s: %w", logOutput, err)
			}

			// Validate the cleaned path matches the resolved absolute path
			if filepath.Clean(absPath) != absPath {
				return fmt.Errorf("invalid log file path: %s", logOutput)
			}

			// Open file for writing (create if not exists, append if exists)
			// Use 0600 permissions (owner read/write only) for security
			logFile, err = os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				return fmt.Errorf("error opening log file %s: %w", logOutput, err)
			}
			output = logFile
		}

		// Initialize logger
		logger.Init(level, logFormat, output)
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close log file if it was opened
		if logFile != nil {
			if err := logFile.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing log file: %v\n", err)
			}
			logFile = nil
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Register adds a subcommand to the root command.
// This function is called by subcommand packages during their init() functions
// to register themselves with the root command.
//
// Parameters:
//   - cmd: The Cobra command to register as a subcommand
func Register(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// GetRootCmd returns the root command instance.
// This is primarily useful for testing, allowing test code to access
// the root command structure.
//
// Returns the root Cobra command instance.
func GetRootCmd() *cobra.Command {
	return rootCmd
}

// Execute executes the root command and handles errors.
// It is the main entry point for the CLI application and should be called
// from the main package. On failure, it exits with code 1.
// Cobra already prints error messages, so this function only handles exit codes.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Configure Cobra to handle errors gracefully
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true

	// Set custom version template to display version, commit, and date information.
	rootCmd.SetVersionTemplate(fmt.Sprintf("mtc %s (%s) %s\n", version.VERSION, version.COMMIT, version.DATE))

	// Set custom help template to show Examples after Flags
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}{{if or .Runnable .HasSubCommands}}{{if .Runnable}}
Usage:
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)

	// Add persistent flags for logging
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "Set the logging level (debug, info, warn, error). Default: warn (only warnings and errors)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Set the logging format (text, json). Default: text")
	rootCmd.PersistentFlags().StringVar(&logOutput, "log-output", "stdout", "Set the log output destination (stdout or a filename). Default: stdout")
	rootCmd.PersistentFlags().CountVarP(&verbose, "verbose", "v", "Enable verbose output: -v for info level, -vv for debug level")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-error output (equivalent to --log-level=error)")
}

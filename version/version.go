// Package version provides build-time version information for the MTC application.
// These variables are set during the build process via linker flags.
package version

var (
	// VERSION is the semantic version of the application (e.g., "1.0.0").
	// Set at build time via -ldflags. Defaults to "dev" if not set.
	VERSION = "dev"

	// COMMIT is the Git commit hash of the build.
	// Set at build time via -ldflags. Defaults to "unknown" if not set.
	COMMIT = "unknown"

	// DATE is the build timestamp in RFC3339 format (e.g., "2024-01-01T12:00:00Z").
	// Set at build time via -ldflags. Defaults to "unknown" if not set.
	DATE = "unknown"
)

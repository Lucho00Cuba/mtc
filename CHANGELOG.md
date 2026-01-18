# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-18

### Added
- Initial release of the MTC CLI tool
- `hash` command for computing Merkle Tree checksums of files and directories
- `diff` command for comparing two directory Merkle trees
- Support for `.mtcignore` and `.gitignore` files
- Custom ignore file support via `--ignore-file` flag
- Exclude patterns support via `--exclude` flag
- Multiple logging levels (debug, info, warn, error)
- JSON and text log formats
- Verbose and quiet output modes
- Cross-platform support (Linux, macOS, Windows)
- Multi-architecture support (amd64, arm64)

### Security
- Symlinks are not traversed and are hashed as leaf nodes to prevent escaping the logical tree

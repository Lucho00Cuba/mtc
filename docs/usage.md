# Usage Guide

This guide covers all commands and options available in MTC, with practical examples for each use case.

## 📋 Table of Contents

- [The `hash` Command](#the-hash-command) - Calculate checksums
- [The `diff` Command](#the-diff-command) - Compare directories
- [The `calc` Command](#the-calc-command) - Verify checksums
- [Global Options](#global-options) - Logging and configuration
- [Exclusion Files](#exclusion-files) - Ignore files and directories

## 🔑 The `hash` Command

The `hash` command calculates the Merkle checksum of a file or directory.

### Basic Syntax

```bash
mtc hash [path]
```

### Basic Examples

```bash
# Calculate hash of a directory
mtc hash ./my-project

# Calculate hash of a file
mtc hash document.pdf

# Calculate hash of current directory
mtc hash .

# Calculate hash of an absolute path
mtc hash /home/user/projects/my-app
```

### Command Output

The `hash` command produces output in the format:

```
[path] ([type]): [hash] (size: [size])
```

Where:
- `[path]` is the path of the file or directory
- `[type]` is `d` for directory or `f` for file
- `[hash]` is the hexadecimal checksum in lowercase
- `[size]` is the total size in human-readable format (B, KB, MB, GB, etc.)

**Example output:**

```
./my-project (d): a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456 (size: 2.5 MB)
document.pdf (f): f6e5d4c3b2a1987654321098765432109876543210fedcba0987654321fedcba09 (size: 1.2 MB)
```

### Exclude Files and Directories

You can exclude specific patterns using the `-e` or `--exclude` option:

```bash
# Exclude a single pattern
mtc hash ./project -e node_modules

# Exclude multiple patterns
mtc hash ./project -e node_modules -e .git -e dist

# Exclude using wildcards
mtc hash ./project -e "*.log" -e "*.tmp"
```

### Using Exclusion Files

MTC automatically loads `.mtcignore` and `.gitignore` from the working directory:

```bash
# MTC automatically uses .mtcignore and .gitignore if they exist
mtc hash ./project
```

You can also specify a custom exclusion file:

```bash
# Use a custom exclusion file
mtc hash ./project --ignore-file=./.mtcignore-custom

# Or with the short form
mtc hash ./project -i ./.mtcignore-custom
```

### Advanced Examples

```bash
# Hash with multiple pattern exclusions and verbose logging
mtc hash ./project -e node_modules -e .git -e dist -v

# Hash with JSON output for automated processing
mtc hash ./project --log-format=json | jq .

# Hash saving logs to file
mtc hash ./project --log-output=mtc.log -vv

# Hash in quiet mode (errors only)
mtc hash ./project --quiet
```

### Using Output in Scripts

The `hash` output is designed to be easily processed:

```bash
# Extract only the hash
HASH=$(mtc hash ./project | awk '{print $3}')
echo "Hash: $HASH"

# Save hash in variable and use in verification
EXPECTED_HASH=$(mtc hash ./project-original | awk '{print $3}')
CURRENT_HASH=$(mtc hash ./project-current | awk '{print $3}')

if [ "$EXPECTED_HASH" = "$CURRENT_HASH" ]; then
    echo "Directories are identical"
else
    echo "Directories differ"
fi
```

## 🔍 The `diff` Command

The `diff` command compares two directories and shows the differences between them.

### Basic Syntax

```bash
mtc diff [pathA] [pathB]
```

### Basic Examples

```bash
# Compare two directories
mtc diff ./backup-before ./backup-after

# Compare current directory with another
mtc diff . ./other-project

# Compare absolute paths
mtc diff /path/to/project-a /path/to/project-b
```

### Command Output

The `diff` command shows one line per difference found:

```
[operation] [path]
```

Where `[operation]` can be:
- `+` - File or directory added in pathB
- `-` - File or directory removed in pathB
- `M` - File modified (content changed)
- `~` - File moved or renamed

**Example output:**

```
M src/main.go
+ src/newfile.go
- src/oldfile.go
+ tests/integration/
M tests/unit/test.go
```

### Examples with Exclusions

```bash
# Compare excluding node_modules and .git
mtc diff ./project-a ./project-b -e node_modules -e .git

# Compare with multiple exclusions
mtc diff ./before ./after -e "*.log" -e "*.tmp" -e dist

# Compare using custom exclusion file
mtc diff ./project-a ./project-b --ignore-file=./.mtcignore
```

### Using Diff in Scripts

```bash
# Check if there are differences
if mtc diff ./backup-before ./backup-after | grep -q .; then
    echo "Differences found"
    mtc diff ./backup-before ./backup-after
else
    echo "Directories are identical"
fi

# Count modified files
MODIFIED=$(mtc diff ./before ./after | grep "^M" | wc -l)
echo "Modified files: $MODIFIED"

# List only new files
mtc diff ./before ./after | grep "^+" | sed 's/^+ //'
```

### Advanced Examples

```bash
# Diff with detailed logging
mtc diff ./project-a ./project-b -vv

# Diff with JSON output
mtc diff ./project-a ./project-b --log-format=json -v

# Diff saving result to file
mtc diff ./before ./after > differences.txt
```

## ✅ The `calc` Command

The `calc` command verifies that a file or directory matches an expected hash.

### Basic Syntax

```bash
mtc calc [path] [hash]
```

### Basic Examples

```bash
# Verify that a directory matches a hash
mtc calc ./my-project a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456

# Verify a file
mtc calc document.pdf f6e5d4c3b2a1987654321098765432109876543210fedcba0987654321fedcba09
```

### Command Output

**If the hash matches:**
```
Hash matches: [hash]
```
The command exits with exit code 0.

**If the hash does NOT match:**
```
Hash mismatch!
Computed: [computed-hash]
Expected: [expected-hash]
```
The command exits with non-zero exit code (1).

### Using Calc in Scripts and CI/CD

The `calc` command is designed to be used in scripts and CI/CD pipelines:

```bash
#!/bin/bash
EXPECTED_HASH="a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"

if mtc calc ./project "$EXPECTED_HASH"; then
    echo "✓ Verification successful"
    exit 0
else
    echo "✗ Verification failed - project has been modified"
    exit 1
fi
```

### Example with Exclusions

```bash
# Verify excluding certain directories
mtc calc ./project abc123... -e node_modules -e .git

# Verify with exclusion file
mtc calc ./project abc123... --ignore-file=./.mtcignore
```

### CI/CD Integration

```yaml
# Example for GitHub Actions
- name: Verify project integrity
  run: |
    EXPECTED_HASH="a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
    mtc calc ./project "$EXPECTED_HASH"
```

## ⚙️ Global Options

All commands share these global options:

### Logging Options

#### Verbose (`-v`, `-vv`)

```bash
# Info level (shows general information)
mtc hash ./project -v

# Debug level (shows detailed information)
mtc hash ./project -vv
```

#### Quiet (`-q`, `--quiet`)

```bash
# Only shows errors
mtc hash ./project --quiet
```

#### Log Level (`--log-level`)

```bash
# Specify level explicitly
mtc hash ./project --log-level=debug
mtc hash ./project --log-level=info
mtc hash ./project --log-level=warn
mtc hash ./project --log-level=error
```

**Available levels:**
- `debug` - Very detailed information (useful for debugging)
- `info` - General information about the process
- `warn` - Warnings (default level)
- `error` - Errors only

#### Log Format (`--log-format`)

```bash
# Text format (default)
mtc hash ./project --log-format=text

# JSON format (useful for automated processing)
mtc hash ./project --log-format=json
```

**Example JSON output:**

```json
{"level":"info","msg":"Starting hash computation","path":"./project","command":"hash"}
{"level":"info","msg":"Hash computation completed","duration":"1.234s","hash":"abc123...","size":"2.5 MB"}
```

#### Log Output (`--log-output`)

```bash
# Output to stdout (default)
mtc hash ./project --log-output=stdout

# Output to file
mtc hash ./project --log-output=mtc.log

# Combine with verbose for detailed logs
mtc hash ./project --log-output=mtc.log -vv
```

### Other Global Options

```bash
# Show version
mtc --version

# Show help
mtc --help

# Show help for a specific command
mtc hash --help
mtc diff --help
mtc calc --help
```

## 📁 Exclusion Files

MTC supports exclusion files similar to `.gitignore`. See the [Advanced Topics](./advanced.md) section for more details on patterns and syntax.

### Automatic Files

MTC automatically loads and uses:
- `.mtcignore` - MTC-specific file (highest priority)
- `.gitignore` - Git file (if it exists)

These files are searched from the current working directory upward.

### Custom File

```bash
# Use a custom exclusion file
mtc hash ./project --ignore-file=./.mtcignore-custom
```

### Priority

1. Custom file (`--ignore-file`) - **Highest priority**
2. Command-line patterns (`-e`, `--exclude`)
3. `.mtcignore`
4. `.gitignore` - **Lowest priority**

## 💡 Tips and Best Practices

### 1. Use Consistent Exclusions

For accurate comparisons, use the same exclusions:

```bash
# Calculate hash with exclusions
HASH=$(mtc hash ./project -e node_modules -e .git | awk '{print $3}')

# Verify with the same exclusions
mtc calc ./project "$HASH" -e node_modules -e .git
```

### 2. Save Hashes for Future Reference

```bash
# Save hash to file
mtc hash ./project > project.hash

# Read and use later
EXPECTED_HASH=$(awk '{print $3}' project.hash)
mtc calc ./project "$EXPECTED_HASH"
```

### 3. Use JSON for Automated Processing

```bash
# Get hash in JSON format
HASH=$(mtc hash ./project --log-format=json -v | \
  jq -r 'select(.msg=="Hash computation completed") | .hash')

echo "Hash: $HASH"
```

### 4. Verify Integrity in Pipelines

```bash
# In a CI/CD script
EXPECTED_HASH="abc123..."
if ! mtc calc ./artifacts "$EXPECTED_HASH" --quiet; then
    echo "ERROR: Artifacts do not match expected hash"
    exit 1
fi
```

### 5. Compare Multiple Versions

```bash
# Create hashes for multiple versions
mtc hash ./v1.0 > v1.0.hash
mtc hash ./v1.1 > v1.1.hash
mtc hash ./v1.2 > v1.2.hash

# Compare versions
mtc diff ./v1.0 ./v1.1
mtc diff ./v1.1 ./v1.2
```

## 🎯 Next Steps

- Explore [Real-World Use Cases](./use-cases.md) to see practical examples
- Learn about [Advanced Topics](./advanced.md) for complex configurations
- Consult the [Installation Guide](./installation.md) if you need help with installation

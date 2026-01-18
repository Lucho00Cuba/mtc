# Advanced Topics

This guide covers advanced topics in MTC, including detailed configuration, exclusion files, best practices, and optimization.

## 📋 Table of Contents

- [Detailed Exclusion Files](#detailed-exclusion-files)
- [Exclusion Patterns](#exclusion-patterns)
- [Best Practices](#best-practices)
- [Optimization and Performance](#optimization-and-performance)
- [Integration with Other Tools](#integration-with-other-tools)
- [Advanced Troubleshooting](#advanced-troubleshooting)

## 📁 Detailed Exclusion Files

MTC supports exclusion files with syntax similar to `.gitignore`, allowing you to exclude files and directories from hash calculation.

### Supported Files

MTC automatically searches for and loads these files (in priority order):

1. **`.mtcignore`** - MTC-specific file (highest priority)
2. **`.gitignore`** - Git file (if it exists)

These files are searched from the current working directory upward in the directory hierarchy.

### Exclusion File Location

Exclusion files are searched from the **current working directory** (where you execute the command), not from the directory you're hashing. This allows having a single exclusion file at the project root.

**Example:**

```bash
# Directory structure
project/
├── .mtcignore
├── src/
│   └── main.go
└── subdirectory/
    └── file.txt

# From project root
cd /path/to/project
mtc hash ./subdirectory  # .mtcignore is loaded from /path/to/project
```

### Custom File

You can specify a custom exclusion file with `--ignore-file`:

```bash
mtc hash ./project --ignore-file=./.mtcignore-custom
```

The custom file has **highest priority** over `.mtcignore` and `.gitignore`.

## 🎯 Exclusion Patterns

### Basic Syntax

Patterns in exclusion files follow syntax similar to `.gitignore`:

#### Exact Matches

```
node_modules
.git
dist
```

#### Directory Matches

Add `/` at the end to match only directories:

```
node_modules/    # Only directories named node_modules
build/           # Only directories named build
```

#### Glob Patterns

Use `*` to match any sequence of characters and `?` for a single character:

```
*.log            # All .log files
*.tmp
test-*.txt       # Files starting with test- and ending in .txt
file?.txt        # file1.txt, file2.txt, etc.
```

#### Recursive Patterns

Use `**` to match any number of directories:

```
**/node_modules  # node_modules at any level
**/*.log         # All .log files at any level
**/build/**      # Everything inside build directories at any level
```

#### Negation

Use `!` at the start to negate a pattern (include files that were excluded):

```
# Exclude all .log
*.log

# But include important.log
!important.log

# Exclude all files in temp/
temp/*

# But include temp/keep-this.txt
!temp/keep-this.txt
```

### `.mtcignore` File Examples

#### For Node.js Project

```
# Dependencias
node_modules/
package-lock.json
yarn.lock

# Build outputs
dist/
build/
*.min.js
*.min.css

# Logs
*.log
logs/

# Archivos temporales
*.tmp
*.temp
.cache/

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Pero incluir archivos importantes
!dist/important.js
```

#### For Python Project

```
# Virtual environments
venv/
env/
.venv/

# Python cache
__pycache__/
*.py[cod]
*.pyc
*.pyo
*.pyd
.Python

# Distribution / packaging
dist/
build/
*.egg-info/

# Testing
.pytest_cache/
.coverage
htmlcov/

# Jupyter
.ipynb_checkpoints/

# Logs
*.log
```

#### For Go Project

```
# Binarios
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Directorio de distribución
dist/

# Dependencias de prueba
vendor/

# IDE
.vscode/
.idea/

# Logs
*.log
```

### Pattern Priority

Patterns are evaluated in this order (from highest to lowest priority):

1. **Custom file** (`--ignore-file`)
2. **Command-line patterns** (`-e`, `--exclude`)
3. **`.mtcignore`**
4. **`.gitignore`**

Within each file, patterns are evaluated in order, and negation patterns (`!`) can override previous exclusions.

### Pattern Usage Examples

```bash
# Exclude node_modules and .git using command line
mtc hash ./project -e node_modules -e .git

# Exclude using .mtcignore file
# (create .mtcignore with: node_modules and .git)
mtc hash ./project

# Combine: exclude node_modules from CLI and use .mtcignore for the rest
mtc hash ./project -e node_modules

# Use custom file with highest priority
mtc hash ./project --ignore-file=./.mtcignore-production
```

## 💡 Best Practices

### 1. Consistency in Exclusions

**Always use the same exclusions** when calculating and verifying hashes:

```bash
# ❌ INCORRECT - different exclusions
HASH=$(mtc hash ./project -e node_modules | awk '{print $3}')
mtc calc ./project "$HASH"  # Does not include node_modules exclusion

# ✅ CORRECT - same exclusions
HASH=$(mtc hash ./project -e node_modules | awk '{print $3}')
mtc calc ./project "$HASH" -e node_modules
```

### 2. Use Exclusion Files Instead of CLI

For permanent exclusions, use `.mtcignore` files instead of command-line options:

```bash
# ❌ Less maintainable
mtc hash ./project -e node_modules -e .git -e dist -e "*.log"

# ✅ Better - use .mtcignore
# .mtcignore contains:
# node_modules
# .git
# dist
# *.log
mtc hash ./project
```

### 3. Version Exclusion Files

Include `.mtcignore` in your version control so all team members use the same exclusions:

```bash
git add .mtcignore
git commit -m "Add .mtcignore for consistent hashing"
```

### 4. Document Exclusions

Comment your exclusion files so others understand why certain files are excluded:

```
# Exclude dependencies - installed from package.json
node_modules/

# Exclude builds - generated from source code
dist/
build/

# Exclude logs - may vary between runs
*.log
logs/
```

### 5. Validate Hashes in Pipelines

Always validate hashes in your CI/CD pipelines:

```yaml
- name: Verify artifact integrity
  run: |
    if ! mtc calc ./artifacts "$EXPECTED_HASH"; then
      echo "Artifact verification failed"
      exit 1
    fi
```

### 6. Store Reference Hashes

Save reference hashes in a safe and versioned location:

```bash
# Save version hash
VERSION="1.2.3"
HASH=$(mtc hash ./project | awk '{print $3}')
echo "$VERSION|$HASH" >> hashes.txt

# Verify specific version
EXPECTED_HASH=$(grep "^$VERSION|" hashes.txt | cut -d'|' -f2)
mtc calc ./project "$EXPECTED_HASH"
```

## ⚡ Optimization and Performance

### Large Directories

For very large directories, consider excluding files you don't need to verify:

```bash
# Exclude large files that don't change frequently
mtc hash ./project -e "*.mp4" -e "*.iso" -e "*.zip"
```

### Multiple Verifications

If you need to verify multiple directories, use scripts instead of individual commands:

```bash
#!/bin/bash
# Verify multiple directories efficiently

DIRS=(
    "/path/dir1"
    "/path/dir2"
    "/path/dir3"
)

for dir in "${DIRS[@]}"; do
    echo "Verifying: $dir"
    HASH=$(mtc hash "$dir" --quiet | awk '{print $3}')
    echo "  Hash: $HASH"
done
```

### Logging in Production

In production, use quiet mode or log to file:

```bash
# Quiet mode (errors only)
mtc hash ./project --quiet

# Logging to file
mtc hash ./project --log-output=/var/log/mtc.log --log-level=warn
```

### Parallel Processing

Although MTC is already efficient internally, you can process multiple directories in parallel:

```bash
#!/bin/bash
# Process multiple directories in parallel

hash_dir() {
    dir=$1
    mtc hash "$dir" --quiet | awk '{print $3}'
}

export -f hash_dir

# Process in parallel (maximum 4 jobs)
DIRS=(/path/dir1 /path/dir2 /path/dir3 /path/dir4)
parallel -j 4 hash_dir ::: "${DIRS[@]}"
```

## 🔗 Integration with Other Tools

### Git Integration

You can use MTC to verify that files in Git have not been modified locally:

```bash
#!/bin/bash
# Verify that working directory matches last commit

# Hash of working directory (excluding .git)
WORKING_HASH=$(mtc hash . -e .git | awk '{print $3}')

# Hash of last commit (using git archive)
git archive HEAD | tar -x -C /tmp/last-commit
COMMIT_HASH=$(mtc hash /tmp/last-commit | awk '{print $3}')

if [ "$WORKING_HASH" = "$COMMIT_HASH" ]; then
    echo "Working directory matches last commit"
else
    echo "Working directory differs from last commit"
    mtc diff /tmp/last-commit .
fi
```

### Docker Integration

Verify Docker image integrity:

```bash
#!/bin/bash
# Extract and verify Docker image content

IMAGE="my-image:latest"
EXTRACT_DIR="/tmp/docker-extract"

# Extract image content
docker create --name temp-container "$IMAGE"
docker cp temp-container:/ "$EXTRACT_DIR"
docker rm temp-container

# Calculate hash
HASH=$(mtc hash "$EXTRACT_DIR" | awk '{print $3}')
echo "Image content hash: $HASH"
```

### rsync Integration

Verify that rsync synchronized correctly:

```bash
#!/bin/bash
# Synchronize and verify

SOURCE="/mnt/source"
DEST="/mnt/dest"

# Synchronize
rsync -av "$SOURCE/" "$DEST/"

# Verify
SOURCE_HASH=$(mtc hash "$SOURCE" | awk '{print $3}')
DEST_HASH=$(mtc hash "$DEST" | awk '{print $3}')

if [ "$SOURCE_HASH" = "$DEST_HASH" ]; then
    echo "✓ Synchronization verified"
else
    echo "✗ Synchronization error"
    mtc diff "$SOURCE" "$DEST"
fi
```

### Make Integration

```makefile
# Makefile with hash verification

.PHONY: hash verify

# Calculate project hash
hash:
	@echo "Calculating project hash..."
	@mtc hash . > .project-hash
	@cat .project-hash

# Verify that project has not changed
verify:
	@if [ -f .project-hash ]; then \
		EXPECTED=$$(awk '{print $$3}' .project-hash); \
		mtc calc . "$$EXPECTED" || exit 1; \
	else \
		echo "No hash file found. Run 'make hash' first."; \
		exit 1; \
	fi

# Build with verification
build: verify
	@echo "Building..."
	# Your build command here
```

## 🔧 Advanced Troubleshooting

### Different Hash on Same Platform

If you get different hashes for the same directory:

1. **Check exclusions**: Make sure you're using the same exclusions
2. **Check hidden files**: Some files may not be visible
3. **Check permissions**: Permissions don't affect the hash, but symbolic links do

```bash
# Check exclusions
mtc hash ./project -vv  # Debug mode to see what's excluded

# List all files (including hidden)
find ./project -type f | sort

# Check symbolic links
find ./project -type l
```

### Different Hash Between Platforms

MTC is designed to be deterministic across platforms. If you get different hashes:

1. **Check path separators**: MTC normalizes paths internally
2. **Check file encoding**: Some files may have different encodings
3. **Check platform-specific files**: Exclude files like `.DS_Store` (macOS) or `Thumbs.db` (Windows)

```bash
# Exclude platform-specific files
mtc hash ./project -e .DS_Store -e Thumbs.db -e "*.swp"
```

### Slow Performance

If MTC is slow on large directories:

1. **Exclude unnecessary large files**:
```bash
mtc hash ./project -e "*.mp4" -e "*.iso" -e "*.zip" -e "*.tar.gz"
```

2. **Use quiet mode** to reduce logging overhead:
```bash
mtc hash ./project --quiet
```

3. **Check disk I/O**: Performance depends mainly on disk read speed

### Exclusion File Issues

If exclusion files don't work as expected:

1. **Check location**: Files are searched from the working directory, not from the hashed directory
2. **Check syntax**: Use `-vv` to see what patterns are loaded
3. **Check priority**: The custom file has highest priority

```bash
# Debug: see what patterns are loaded
mtc hash ./project -vv 2>&1 | grep -i "ignore\|exclude"

# Test specific pattern
mtc hash ./project -e "test-pattern" -vv
```

## 🎯 Next Steps

- Review the [Usage Guide](./usage.md) for basic commands
- Explore [Real-World Use Cases](./use-cases.md) for practical examples
- Consult the [main README](../README.md) for general information

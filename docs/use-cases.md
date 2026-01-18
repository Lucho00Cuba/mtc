# Real-World Use Cases

This guide presents practical and real-world use cases where MTC can be very useful. Each case includes concrete examples and scripts that you can adapt to your needs.

## 📋 Table of Contents

- [Backup Verification](#backup-verification)
- [Configuration Drift Detection](#configuration-drift-detection)
- [CI/CD Artifact Validation](#cicd-artifact-validation)
- [Change Auditing](#change-auditing)
- [CI Reproducibility](#ci-reproducibility)
- [Deployment Validation](#deployment-validation)
- [Directory Synchronization](#directory-synchronization)
- [Configuration File Integrity](#configuration-file-integrity)

## 💾 Backup Verification

### Problem

You need to verify that your backups were created correctly and were not corrupted during storage or transfer.

### Solution

Calculate the hash of the original directory before backup and verify after restoring.

### Complete Example

```bash
#!/bin/bash
# Backup verification script

BACKUP_SOURCE="/home/user/important-data"
BACKUP_DEST="/mnt/backup/important-data-$(date +%Y%m%d)"
EXPECTED_HASH_FILE="/mnt/backup/hashes/important-data-$(date +%Y%m%d).hash"

# 1. Calculate hash of original directory
echo "Calculating hash of original directory..."
ORIGINAL_HASH=$(mtc hash "$BACKUP_SOURCE" | awk '{print $3}')
echo "Original hash: $ORIGINAL_HASH"

# 2. Save hash for future reference
mkdir -p "$(dirname "$EXPECTED_HASH_FILE")"
echo "$ORIGINAL_HASH" > "$EXPECTED_HASH_FILE"

# 3. Create backup (example with rsync)
echo "Creating backup..."
rsync -av "$BACKUP_SOURCE/" "$BACKUP_DEST/"

# 4. Verify backup integrity
echo "Verifying backup integrity..."
if mtc calc "$BACKUP_DEST" "$ORIGINAL_HASH"; then
    echo "✓ Backup verified successfully"
    exit 0
else
    echo "✗ ERROR: Backup does not match original"
    exit 1
fi
```

### Periodic Verification

```bash
#!/bin/bash
# Verify all stored backups

BACKUP_DIR="/mnt/backup"
HASHES_DIR="$BACKUP_DIR/hashes"

for hash_file in "$HASHES_DIR"/*.hash; do
    if [ -f "$hash_file" ]; then
        backup_name=$(basename "$hash_file" .hash)
        backup_path="$BACKUP_DIR/$backup_name"
        expected_hash=$(cat "$hash_file")

        if [ -d "$backup_path" ]; then
            echo "Verifying: $backup_name"
            if mtc calc "$backup_path" "$expected_hash" --quiet; then
                echo "  ✓ OK"
            else
                echo "  ✗ CORRUPT"
            fi
        fi
    fi
done
```

## 🔄 Configuration Drift Detection

### Problem

You need to detect when configuration files in production have been modified without authorization (configuration drift).

### Solution

Maintain reference hashes of known configurations and compare periodically.

### Complete Example

```bash
#!/bin/bash
# Configuration drift detection script

CONFIG_DIR="/etc/my-application"
REFERENCE_HASH_FILE="/var/lib/my-app/config.hash"
ALERT_EMAIL="admin@example.com"

# Calculate current hash
CURRENT_HASH=$(mtc hash "$CONFIG_DIR" | awk '{print $3}')

# Check if reference hash exists
if [ -f "$REFERENCE_HASH_FILE" ]; then
    REFERENCE_HASH=$(cat "$REFERENCE_HASH_FILE")

    if [ "$CURRENT_HASH" != "$REFERENCE_HASH" ]; then
        echo "ALERT: Configuration modified"
        echo "Reference hash: $REFERENCE_HASH"
        echo "Current hash: $CURRENT_HASH"

        # Get detailed differences
        echo ""
        echo "Differences detected:"
        mtc diff "$CONFIG_DIR" "/var/lib/my-app/config-reference"

        # Send alert (example)
        # mail -s "Alert: Configuration modified" "$ALERT_EMAIL" < /tmp/drift-alert.txt

        exit 1
    else
        echo "Configuration unchanged"
        exit 0
    fi
else
    # First run: save reference hash
    echo "Creating initial reference hash..."
    echo "$CURRENT_HASH" > "$REFERENCE_HASH_FILE"
    cp -r "$CONFIG_DIR" "/var/lib/my-app/config-reference"
    echo "Reference hash saved: $CURRENT_HASH"
    exit 0
fi
```

### Cron Integration

```bash
# /etc/cron.d/config-drift-check
# Verify configuration every hour
0 * * * * root /usr/local/bin/check-config-drift.sh
```

## 🚀 CI/CD Artifact Validation

### Problem

You need to verify that artifacts generated in your CI/CD pipeline have not been modified or corrupted.

### Solution

Calculate hash of artifacts after building and verify before deployment.

### Ejemplo para GitHub Actions

```yaml
name: Build and Verify Artifacts

on:
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install MTC
        run: |
          curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-linux-amd64
          chmod +x mtc-linux-amd64
          sudo mv mtc-linux-amd64 /usr/local/bin/mtc

      - name: Build artifacts
        run: |
          make build
          mkdir -p artifacts
          cp dist/* artifacts/

      - name: Calculate artifact hash
        id: hash
        run: |
          HASH=$(mtc hash ./artifacts | awk '{print $3}')
          echo "hash=$HASH" >> $GITHUB_OUTPUT
          echo "Artifact hash: $HASH"

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: build-artifacts
          path: artifacts/

      - name: Upload hash
        uses: actions/upload-artifact@v3
        with:
          name: artifact-hash
          path: |
            hash.txt
        env:
          HASH: ${{ steps.hash.outputs.hash }}
        run: |
          echo "$HASH" > hash.txt

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download artifacts
        uses: actions/download-artifact@v3
        with:
          name: build-artifacts

      - name: Download hash
        uses: actions/download-artifact@v3
        with:
          name: artifact-hash

      - name: Install MTC
        run: |
          curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-linux-amd64
          chmod +x mtc-linux-amd64
          sudo mv mtc-linux-amd64 /usr/local/bin/mtc

      - name: Verify artifact integrity
        run: |
          EXPECTED_HASH=$(cat hash.txt)
          if mtc calc ./artifacts "$EXPECTED_HASH"; then
            echo "✓ Artifacts verified"
          else
            echo "✗ Artifact verification failed"
            exit 1
          fi

      - name: Deploy
        run: |
          # Your deployment logic here
          echo "Deploying verified artifacts..."
```

### Example for GitLab CI

```yaml
stages:
  - build
  - verify
  - deploy

build:
  stage: build
  script:
    - make build
    - mkdir -p artifacts
    - cp dist/* artifacts/
    - |
      # Calculate and save hash
      HASH=$(mtc hash ./artifacts | awk '{print $3}')
      echo "$HASH" > artifact.hash
      echo "Artifact hash: $HASH"
  artifacts:
    paths:
      - artifacts/
      - artifact.hash

verify:
  stage: verify
  script:
    - |
      EXPECTED_HASH=$(cat artifact.hash)
      if mtc calc ./artifacts "$EXPECTED_HASH"; then
        echo "✓ Verification successful"
      else
        echo "✗ Verification failed"
        exit 1
      fi

deploy:
  stage: deploy
  script:
    - echo "Deploying verified artifacts..."
    # Your deployment logic
```

## 📊 Change Auditing

### Problem

You need to maintain a record of changes in critical directories for auditing and compliance.

### Solution

Calculate hashes periodically and compare with previous hashes to detect changes.

### Audit Script

```bash
#!/bin/bash
# Change auditing system

AUDIT_DIRS=(
    "/etc"
    "/opt/my-application/config"
    "/var/lib/my-app/data"
)

AUDIT_LOG="/var/log/mtc-audit.log"
HASH_STORAGE="/var/lib/mtc-audit/hashes"
DATE=$(date +%Y%m%d-%H%M%S)

mkdir -p "$HASH_STORAGE"

for dir in "${AUDIT_DIRS[@]}"; do
    if [ ! -d "$dir" ]; then
        echo "[$DATE] WARNING: Directory not found: $dir" >> "$AUDIT_LOG"
        continue
    fi

    # Calculate current hash
    CURRENT_HASH=$(mtc hash "$dir" | awk '{print $3}')
    HASH_FILE="$HASH_STORAGE/$(echo "$dir" | tr '/' '_')-latest.hash"

    if [ -f "$HASH_FILE" ]; then
        PREVIOUS_HASH=$(cat "$HASH_FILE")

        if [ "$CURRENT_HASH" != "$PREVIOUS_HASH" ]; then
            echo "[$DATE] CHANGE DETECTED: $dir" >> "$AUDIT_LOG"
            echo "  Previous: $PREVIOUS_HASH" >> "$AUDIT_LOG"
            echo "  Current:  $CURRENT_HASH" >> "$AUDIT_LOG"

            # Save historical hash
            cp "$HASH_FILE" "$HASH_STORAGE/$(echo "$dir" | tr '/' '_')-$DATE.hash"
        else
            echo "[$DATE] NO CHANGE: $dir" >> "$AUDIT_LOG"
        fi
    else
        echo "[$DATE] INITIAL HASH: $dir -> $CURRENT_HASH" >> "$AUDIT_LOG"
    fi

    # Save current hash
    echo "$CURRENT_HASH" > "$HASH_FILE"
done
```

### Generate Change Report

```bash
#!/bin/bash
# Generate change report from logs

AUDIT_LOG="/var/log/mtc-audit.log"
REPORT_FILE="/tmp/audit-report-$(date +%Y%m%d).txt"

echo "Audit Report - $(date)" > "$REPORT_FILE"
echo "=================================" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Count changes
CHANGES=$(grep "CHANGE DETECTED" "$AUDIT_LOG" | wc -l)
echo "Total changes detected: $CHANGES" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# List recent changes
echo "Recent changes:" >> "$REPORT_FILE"
grep "CHANGE DETECTED" "$AUDIT_LOG" | tail -20 >> "$REPORT_FILE"

cat "$REPORT_FILE"
```

## 🔬 CI Reproducibility

### Problem

You need to ensure that builds are reproducible and that the same inputs produce the same outputs.

### Solution

Calculate hash of dependencies and source code, and verify that successive builds with the same hash produce the same artifacts.

### Reproducibility Script

```bash
#!/bin/bash
# Verify build reproducibility

SOURCE_DIR="./src"
DEPS_DIR="./vendor"
BUILD_DIR="./build"

# Calculate hash of inputs (source code and dependencies)
echo "Calculating input hash..."
SOURCE_HASH=$(mtc hash "$SOURCE_DIR" | awk '{print $3}')
DEPS_HASH=$(mtc hash "$DEPS_DIR" | awk '{print $3}')

INPUT_HASH="${SOURCE_HASH}-${DEPS_HASH}"
echo "Input hash: $INPUT_HASH"

# Build
echo "Building..."
make clean
make build

# Calculate hash of outputs
OUTPUT_HASH=$(mtc hash "$BUILD_DIR" | awk '{print $3}')
echo "Output hash: $OUTPUT_HASH"

# Save input-output relationship
echo "$INPUT_HASH|$OUTPUT_HASH" >> build-history.txt

# Check if this input was built before
if grep -q "^$INPUT_HASH|" build-history.txt; then
    PREVIOUS_OUTPUT=$(grep "^$INPUT_HASH|" build-history.txt | head -1 | cut -d'|' -f2)

    if [ "$OUTPUT_HASH" = "$PREVIOUS_OUTPUT" ]; then
        echo "✓ Build reproducible"
    else
        echo "✗ Build NOT reproducible"
        echo "  Previous output: $PREVIOUS_OUTPUT"
        echo "  Current output:  $OUTPUT_HASH"
        exit 1
    fi
else
    echo "First build with these inputs"
fi
```

## 🚢 Deployment Validation

### Problem

You need to verify that files deployed in production exactly match the built artifacts.

### Solution

Compare the hash of the deployed directory with the hash of the original artifact.

### Post-Deployment Validation Script

```bash
#!/bin/bash
# Validate deployment

ARTIFACT_DIR="./artifacts"
DEPLOY_DIR="/var/www/my-application"
EXPECTED_HASH_FILE="./artifact.hash"

# Read expected hash
if [ ! -f "$EXPECTED_HASH_FILE" ]; then
    echo "ERROR: Hash file not found: $EXPECTED_HASH_FILE"
    exit 1
fi

EXPECTED_HASH=$(cat "$EXPECTED_HASH_FILE")
echo "Expected hash: $EXPECTED_HASH"

# Verify deployment
echo "Verifying deployment in $DEPLOY_DIR..."
if mtc calc "$DEPLOY_DIR" "$EXPECTED_HASH"; then
    echo "✓ Deployment verified successfully"
    exit 0
else
    echo "✗ ERROR: Deployment does not match artifacts"

    # Show differences if there's a reference directory
    if [ -d "$ARTIFACT_DIR" ]; then
        echo ""
        echo "Differences detected:"
        mtc diff "$ARTIFACT_DIR" "$DEPLOY_DIR"
    fi

    exit 1
fi
```

## 🔄 Directory Synchronization

### Problem

You need to verify that two directories are correctly synchronized.

### Solution

Compare hashes of both directories to verify synchronization.

### Synchronization Verification Script

```bash
#!/bin/bash
# Verify synchronization between directories

SOURCE_DIR="/mnt/source"
DEST_DIR="/mnt/destination"

echo "Verifying synchronization..."
echo "Source: $SOURCE_DIR"
echo "Destination: $DEST_DIR"

# Calculate hashes
SOURCE_HASH=$(mtc hash "$SOURCE_DIR" | awk '{print $3}')
DEST_HASH=$(mtc hash "$DEST_DIR" | awk '{print $3}')

if [ "$SOURCE_HASH" = "$DEST_HASH" ]; then
    echo "✓ Directories synchronized"
    echo "Hash: $SOURCE_HASH"
    exit 0
else
    echo "✗ Directories NOT synchronized"
    echo "Source hash:      $SOURCE_HASH"
    echo "Destination hash: $DEST_HASH"

    echo ""
    echo "Differences:"
    mtc diff "$SOURCE_DIR" "$DEST_DIR"

    exit 1
fi
```

## 🔐 Configuration File Integrity

### Problem

You need to detect unauthorized changes in sensitive configuration files.

### Solution

Maintain reference hashes and verify periodically.

### Configuration Monitoring Script

```bash
#!/bin/bash
# Monitor changes in sensitive configuration

CONFIG_FILES=(
    "/etc/ssh/sshd_config"
    "/etc/sudoers"
    "/etc/passwd"
    "/etc/shadow"
)

CONFIG_DIR="/etc/my-application"
HASH_STORAGE="/var/lib/config-integrity"

mkdir -p "$HASH_STORAGE"

# Verify individual files
for file in "${CONFIG_FILES[@]}"; do
    if [ -f "$file" ]; then
        CURRENT_HASH=$(mtc hash "$file" | awk '{print $3}')
        HASH_FILE="$HASH_STORAGE/$(basename "$file").hash"

        if [ -f "$HASH_FILE" ]; then
            EXPECTED_HASH=$(cat "$HASH_FILE")
            if [ "$CURRENT_HASH" != "$EXPECTED_HASH" ]; then
                echo "ALERT: $file has been modified"
                echo "  Previous hash: $EXPECTED_HASH"
                echo "  Current hash:  $CURRENT_HASH"
            fi
        else
            echo "Creating reference hash for $file"
            echo "$CURRENT_HASH" > "$HASH_FILE"
        fi
    fi
done

# Verify complete configuration directory
if [ -d "$CONFIG_DIR" ]; then
    CURRENT_HASH=$(mtc hash "$CONFIG_DIR" | awk '{print $3}')
    HASH_FILE="$HASH_STORAGE/config-dir.hash"

    if [ -f "$HASH_FILE" ]; then
        EXPECTED_HASH=$(cat "$HASH_FILE")
        if [ "$CURRENT_HASH" != "$EXPECTED_HASH" ]; then
            echo "ALERT: Configuration directory modified"
            echo "  Previous hash: $EXPECTED_HASH"
            echo "  Current hash:  $CURRENT_HASH"

            # Show what changed
            if [ -d "/var/lib/config-integrity/reference" ]; then
                echo "Changes detected:"
                mtc diff "/var/lib/config-integrity/reference" "$CONFIG_DIR"
            fi
        fi
    else
        echo "Creating reference hash for $CONFIG_DIR"
        echo "$CURRENT_HASH" > "$HASH_FILE"
        cp -r "$CONFIG_DIR" "/var/lib/config-integrity/reference"
    fi
fi
```

## 🎯 Next Steps

- Review the [Usage Guide](./usage.md) for more details on commands
- Consult [Advanced Topics](./advanced.md) for more complex configurations
- Adapt these examples to your specific needs

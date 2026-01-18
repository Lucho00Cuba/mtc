# Installation Guide

This guide will help you install MTC on your system. MTC is available for Linux, macOS, and Windows.

## 📋 Prerequisites

- **Operating System**: Linux, macOS (Darwin), or Windows
- **Architecture**: AMD64 (x86_64) or ARM64
- **No additional dependencies required** - MTC is a static binary

## 🚀 Installation Methods

### Option 1: Download Pre-built Binaries (Recommended)

Pre-built binaries are available in the [Releases](https://github.com/lucho00cuba/mtc/releases) section on GitHub.

#### Linux

```bash
# Download the latest version for Linux AMD64
curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-linux-amd64

# Make executable
chmod +x mtc-linux-amd64

# Move to a directory in your PATH (optional but recommended)
sudo mv mtc-linux-amd64 /usr/local/bin/mtc

# Verify installation
mtc --version
```

For Linux ARM64:

```bash
curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-linux-arm64
chmod +x mtc-linux-arm64
sudo mv mtc-linux-arm64 /usr/local/bin/mtc
```

#### macOS

```bash
# For macOS with Intel processor (AMD64)
curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-darwin-amd64
chmod +x mtc-darwin-amd64
sudo mv mtc-darwin-amd64 /usr/local/bin/mtc

# For macOS with Apple Silicon (ARM64)
curl -LO https://github.com/lucho00cuba/mtc/releases/latest/download/mtc-darwin-arm64
chmod +x mtc-darwin-arm64
sudo mv mtc-darwin-arm64 /usr/local/bin/mtc

# Verify installation
mtc --version
```

#### Windows

1. Download `mtc-windows-amd64.exe` from [Releases](https://github.com/lucho00cuba/mtc/releases)
2. Rename the file to `mtc.exe`
3. Place it in a directory that's in your PATH (e.g., `C:\Program Files\mtc\`)
4. Open PowerShell or CMD and verify:

```powershell
mtc --version
```

### Option 2: Install from Source with Go Install

If you have Go 1.24+ installed:

```bash
go install github.com/lucho00cuba/mtc@latest
```

This will install the `mtc` binary into your `$GOPATH/bin` or `$HOME/go/bin` directory (if `$GOPATH` is not set).

Make sure this directory is in your PATH:

```bash
# Add to PATH (Linux/macOS)
export PATH=$PATH:$(go env GOPATH)/bin

# Or if using the default directory
export PATH=$PATH:$HOME/go/bin
```

### Option 3: Build from Source

If you prefer to build from source:

```bash
# Clone the repository
git clone https://github.com/lucho00cuba/mtc.git
cd mtc

# Build (requires Go 1.24+)
make build

# The binary will be in dist/
# For your specific platform:
./dist/mtc-linux-amd64 --version  # Linux
./dist/mtc-darwin-amd64 --version # macOS Intel
./dist/mtc-darwin-arm64 --version # macOS Apple Silicon
./dist/mtc-windows-amd64.exe --version # Windows
```

#### Build for a Specific Platform

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o mtc-linux-amd64

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o mtc-darwin-amd64

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o mtc-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o mtc-windows-amd64.exe
```

## ✅ Verify Installation

After installing, verify that MTC works correctly:

```bash
# Check version
mtc --version

# Check help
mtc --help

# Test with an example directory
mkdir test-dir
echo "Hello, MTC!" > test-dir/test.txt
mtc hash test-dir
```

You should see something like:

```
test-dir (d): a1b2c3d4e5f6... (size: 12 B)
```

## 🔧 Troubleshooting

### Command 'mtc' not found

**Linux/macOS:**
- Make sure the binary is in a directory that's in your PATH
- Check your PATH: `echo $PATH`
- Add the directory to PATH if necessary

**Windows:**
- Verify that the directory where `mtc.exe` is located is in your system PATH
- You can check in PowerShell: `$env:PATH`

### Permission error

```bash
# Linux/macOS: make the binary executable
chmod +x mtc

# Or if you moved it to /usr/local/bin
sudo chmod +x /usr/local/bin/mtc
```

### Incompatible Go version

MTC requires Go 1.24 or higher. Check your version:

```bash
go version
```

If you need to update Go, visit [golang.org](https://golang.org/dl/).

## 📦 Package Manager Installation

### Homebrew (macOS/Linux)

There is currently no official Homebrew tap, but you can install manually:

```bash
# Download and install
brew install --formula https://raw.githubusercontent.com/lucho00cuba/mtc/main/Formula/mtc.rb
```

Or create your own local tap.

### Scoop (Windows)

```powershell
# Add bucket (if needed)
scoop bucket add extras

# Install (when available)
scoop install mtc
```

## 🔄 Updating MTC

To update to the latest version:

**With Go Install:**
```bash
go install github.com/lucho00cuba/mtc@latest
```

**With pre-built binaries:**
1. Download the new version from [Releases](https://github.com/lucho00cuba/mtc/releases)
2. Replace the old binary with the new one
3. Verify: `mtc --version`

## 🎯 Next Steps

Once MTC is installed, you can:

1. Read the [Usage Guide](./usage.md) to learn the basic commands
2. Explore [Real-World Use Cases](./use-cases.md) to see practical examples
3. Consult [Advanced Topics](./advanced.md) for complex configurations

## 📝 Additional Notes

- MTC is a static binary, no additional libraries required
- The binary is completely portable - you can move it to any location
- No special permissions required to run MTC (except to read the files you want to hash)

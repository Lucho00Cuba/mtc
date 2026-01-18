# MTC

**Merkle Tree Checksum** — A deterministic directory checksum tool that generates directory checksums using Merkle Trees

[![CI](https://github.com/lucho00cuba/mtc/actions/workflows/ci.yml/badge.svg)](https://github.com/lucho00cuba/mtc/actions)
[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-PolyForm%20Noncommercial%201.0.0-blue.svg)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/lucho00cuba/mtc?include_prereleases&sort=semver)](https://github.com/lucho00cuba/mtc/releases)

**MTC** (Merkle Tree Checksum) is a deterministic directory checksum tool that generates directory checksums using Merkle Trees. It recursively computes BLAKE3 hashes for files, builds a Merkle Tree that represents the entire directory structure, and produces a deterministic Merkle checksum that uniquely fingerprints the full state of a directory.

MTC can also compare two directory trees and report exactly what changed — instantly and precisely.

---

- 🔍 Recursive file hashing (BLAKE3)
- 🌳 Deterministic Merkle Tree generation
- ⚖️ Integrity verification via deterministic Merkle checksum comparison
- 📊 Deep diff between directory states
- 🧾 Stable cross-platform output
- 🚀 Extremely fast and memory-efficient
- 🛠️ Single static binary (no runtime dependencies)

---

## 🧩 Why Merkle Trees?

Instead of hashing a directory as a flat blob, MTC builds a hierarchical structure:

- Each file → hashed individually
- Each directory → hash derived from its children
- Merkle checksum → represents the full tree deterministically

This allows:

✔ Fast comparisons
✔ Precise change detection
✔ Integrity validation at scale
✔ Minimal recomputation

---

## 🏛 About MTC

MTC (Merkle Tree Checksum) provides deterministic directory integrity verification through Merkle Tree structures, enabling fast and precise change detection across filesystem hierarchies.

---

⚠️ License Notice
MTC is licensed strictly for non-commercial use.
You may not use MTC within any for-profit organization, internal commercial workflow, proprietary software, SaaS, paid platform, or revenue-generating product without explicit prior written permission from the author.
This restriction also applies to embedding MTC or using it as a dependency within commercial software, services, or internal tooling.

---

## 📦 Installation

### From Source

```bash
git clone https://github.com/lucho00cuba/mtc.git
cd mtc
make build
```

The binary will be available in the `dist/` directory.

### Using Go Install

```bash
go install github.com/lucho00cuba/mtc@latest
```

> This will install the `mtc` CLI binary into your Go bin directory.

### Download Pre-built Binaries

Pre-built binaries for Linux, macOS, and Windows are available in the [Releases](https://github.com/lucho00cuba/mtc/releases) section.

---

## 📚 Documentation

For comprehensive documentation, including detailed installation guides, usage examples, real-world use cases, and advanced topics, see the [Documentation](./docs/README.md) directory.

- **[Installation Guide](./docs/installation.md)** - Detailed installation instructions for all platforms
- **[Usage Guide](./docs/usage.md)** - Complete command reference with examples
- **[Use Cases](./docs/use-cases.md)** - Real-world scenarios and practical examples
- **[Advanced Topics](./docs/advanced.md)** - Advanced configuration and best practices

---

## 🚀 Usage

### Compute Merkle checksum of a directory

```bash
mtc hash ./my-folder
```

### Hash a single file

```bash
mtc hash file.txt
```

### Compare two directory states

```bash
mtc diff ./snapshot-A ./snapshot-B
```

### Logging options

```bash
# Verbose output: -v for info level, -vv for debug level
mtc hash ./dir -v      # info level
mtc hash ./dir -vv     # debug level

# Quiet mode (errors only)
mtc hash ./dir --quiet

# Custom log level
mtc hash ./dir --log-level=debug

# JSON log format
mtc hash ./dir --log-format=json
```

---

## 🧱 Technical Overview

MTC works as follows:

1. Walks the filesystem recursively
2. Hashes files using streaming BLAKE3
3. Sorts directory entries deterministically
4. Builds Merkle nodes bottom-up
5. Produces a deterministic Merkle checksum representing the entire structure

Designed for:

- Large directory trees
- Fast integrity verification
- Low memory footprint
- High concurrency efficiency

---

## 🎯 Use Cases

- Backup verification
- Drift detection
- Artifact integrity checks
- CI reproducibility
- Deployment validation
- Change auditing

---

## ⚡ Performance Philosophy

MTC is optimized to be:

- IO-efficient
- Uses BLAKE3 for highly parallel, SIMD-optimized hashing
- Concurrent but controlled
- Deterministic across platforms
- Minimal memory overhead

Built in Go to deliver a fast, portable binary ideal for DevOps and SRE workflows.

---

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on the process for submitting pull requests.

We have a [Code of Conduct](CODE_OF_CONDUCT.md) that all contributors are expected to follow.

---

## 🔒 Security

For security vulnerabilities, please see our [Security Policy](SECURITY.md).

---

## 📄 License

This project is licensed under PolyForm Noncommercial 1.0.0. See the LICENSE file for full terms.

---

## 🙏 Acknowledgments

- Inspired by content-addressable storage and Merkle-based integrity systems
- Built with [Cobra](https://github.com/spf13/cobra) for CLI functionality

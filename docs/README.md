# MTC Documentation

Welcome to the complete documentation for **MTC** (Merkle Tree Checksum), a CLI tool for generating deterministic directory checksums using Merkle Trees.

## 📚 Documentation Index

- **[Installation](./installation.md)** - Complete guide to install MTC on different platforms
- **[Usage Guide](./usage.md)** - Detailed examples of all commands and options
- **[Real-World Use Cases](./use-cases.md)** - Practical scenarios and real-world examples
- **[Advanced Topics](./advanced.md)** - Advanced configuration, exclusion files, and best practices

## 🚀 Quick Start

If you already have MTC installed, here are the most common commands:

```bash
# Calculate the hash of a directory
mtc hash ./my-project

# Compare two directories
mtc diff ./backup-before ./backup-after

# Verify that a directory matches a hash
mtc calc ./my-project abc123def456...
```

## 🎯 What is MTC?

MTC is a command-line tool written in Go that generates deterministic checksums of entire directories using Merkle Tree structures. Unlike traditional tools that simply concatenate files, MTC builds a hierarchical tree that enables:

- ✅ **Precise change detection** - Identifies exactly which files changed
- ✅ **Fast verification** - Compares checksums without needing to process all files
- ✅ **Guaranteed integrity** - Any change in any file is reflected in the checksum
- ✅ **Deterministic** - The same directory always produces the same hash, regardless of platform

## 🔑 Key Features

- **Recursive hashing** using BLAKE3 (fast and secure)
- **Deterministic Merkle Tree generation**
- **Integrity verification** through checksum comparison
- **Deep diff** between directory states
- **Stable** and cross-platform output
- **Extremely fast** and memory-efficient
- **Single static binary** (no runtime dependencies)

## 📖 Next Steps

1. If you haven't installed MTC yet, start with the [Installation Guide](./installation.md)
2. Learn how to use MTC with the [Usage Guide](./usage.md)
3. Explore [Real-World Use Cases](./use-cases.md) to see how others use MTC
4. Deep dive with [Advanced Topics](./advanced.md) for complex configurations

## 🤝 Contributing

Found an error in the documentation or have suggestions? Please check our [Contributing Guide](../CONTRIBUTING.md) and [Code of Conduct](../CODE_OF_CONDUCT.md).

## 📄 License

MTC is licensed under PolyForm Noncommercial 1.0.0. See the [LICENSE](../LICENSE) file for more details.

---

**Important Note**: MTC is licensed strictly for non-commercial use. You may not use MTC within any for-profit organization, internal commercial workflow, proprietary software, SaaS, paid platform, or revenue-generating product without explicit prior written permission from the author.

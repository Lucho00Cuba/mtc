# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. Do Not Disclose Publicly

**Please do not** open a public GitHub issue for security vulnerabilities. This could put users at risk.

### 2. Report Privately

Please report security vulnerabilities by emailing the maintainers directly or by using GitHub's [private vulnerability reporting](https://github.com/lucho00cuba/mtc/security/advisories/new).

Include the following information:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)
- Your contact information

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity and complexity

### 4. Disclosure

We will:

- Acknowledge receipt of your report
- Keep you informed of our progress
- Credit you in the security advisory (if desired)
- Coordinate public disclosure after a fix is available

## Security Best Practices

When using MTC (Merkle Tree Checksum):

- Always verify checksums of downloaded binaries
- Download from official sources only
- Keep MTC updated to the latest version
- Review file paths before processing
- Be cautious when processing untrusted directories
- Use appropriate file permissions

## Known Security Considerations

### File System Access

MTC reads and processes files from the filesystem. Be aware that:

- MTC will process all files in the specified directory (except those in ignore files)
- Large directories may consume significant resources
- Symlinks are NOT traversed; MTC hashes the link target path as a leaf node to prevent escaping the logical tree and ensure deterministic results
- File permissions are respected

### Input Validation

- MTC normalizes and resolves all paths before processing to guarantee deterministic traversal and prevent unintended escapes via symlinks or relative segments
- Paths are cleaned and evaluated within the logical scope of the requested root before hashing
- Invalid paths result in clear error messages

### Logging

- MTC may log file paths and operations
- Sensitive data should not be stored in file names or paths processed by MTC
- Use appropriate log levels in production environments

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 1.0.1, 1.0.2)
- Documented in CHANGELOG.md
- Tagged with security advisories when appropriate
- Announced through GitHub releases

## Security Checklist for Contributors

When contributing code:

- [ ] Validate all user inputs
- [ ] Sanitize file paths
- [ ] Avoid logging sensitive information
- [ ] Use secure defaults
- [ ] Handle errors securely
- [ ] Review dependencies for vulnerabilities
- [ ] Test security-related changes thoroughly

Thank you for helping keep MTC secure!

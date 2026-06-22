# Security Policy

## Reporting a Vulnerability

We take the security of NeoCode seriously. If you discover a security vulnerability, please report it responsibly.

**DO NOT** open a public GitHub issue for security vulnerabilities.

### How to Report

1. **Email**: Send a report to [WuSuBuDuoMing@users.noreply.github.com](mailto:WuSuBuDuoMing@users.noreply.github.com)
2. **GitHub**: Use [GitHub's private vulnerability reporting](https://github.com/WuSuBuDuoMing/aidev/security/advisories/new)

### What to Include

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

## Supported Versions

| Version | Supported |
|---------|-----------|
| 2.9.x (latest) | Yes |
| 2.8.x | Critical fixes only |
| < 2.2.0 | No |

## Response Time

| Severity | Response Time |
|----------|---------------|
| Critical (RCE, data exfiltration) | 24-48 hours |
| High (auth bypass, data leak) | 3-5 days |
| Medium (information disclosure) | 7 days |
| Low (minor issues) | 14 days |

## Security Measures

### API Key Storage

- API keys are encrypted at rest using AES-256-GCM
- Stored at `~/.neocode/keystore.enc`
- Never written to config files in plaintext
- Storage path is explicitly disclosed on first use

### File System Access

- Write operations are sandboxed to the project directory
- Path traversal is prevented via `filepath.Abs()` validation
- The permission policy engine controls which tools can execute

### Shell Command Execution

- Shell commands use argument arrays (`exec.Command`), not string interpolation
- This prevents shell injection attacks
- Commands have configurable timeouts (default: 30 seconds)

### Network Security

- SSRF protection for web fetch operations
- All HTTP clients have configurable timeouts (default: 5 minutes)
- API keys are transmitted only over HTTPS

### Permission System

- 4 permission modes: Ask, Auto, Plan, Edit
- Plan mode provides read-only analysis with no file modifications
- Per-tool permission rules with wildcard support
- Read-only tools are automatically allowed; write tools require approval

## Responsible Disclosure

We kindly ask that you:

- Allow reasonable time for us to address the vulnerability before public disclosure
- Avoid exploiting the vulnerability beyond what is necessary to demonstrate it
- Do not access or modify data belonging to other users
- Do not perform destructive testing

## Acknowledgments

We appreciate the security research community and will credit reporters (with their permission) in our security advisories.

Thank you for helping keep NeoCode and its users safe!

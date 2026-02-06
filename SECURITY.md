# Security Policy

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report via:
1. **GitHub Security Advisories**: Use the "Report a vulnerability" button on the Security tab
2. **Email**: security@ownu.app (if configured)

### What to include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline:
- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 7 days
- **Resolution**: Based on severity (see below)

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < 1.0   | :x:                |

## Security Update Policy

| Severity | Response Time | Resolution Time |
|----------|---------------|-----------------|
| Critical | 4 hours       | 24 hours        |
| High     | 24 hours      | 7 days          |
| Medium   | 7 days        | 30 days         |
| Low      | 30 days       | 90 days         |

## Security Features

### Authentication
- **WebAuthn/Passkeys**: Phishing-resistant, passwordless authentication
- **No password storage**: Eliminates password-based attacks
- **Recovery phrases**: BIP-39 compatible, hashed before storage

### Encryption
- **AES-256-GCM**: All sensitive data encrypted at rest
- **Per-user keys**: Each user has unique Data Encryption Key
- **Key derivation**: Argon2id for password-based derivation

### API Security
- **Rate limiting**: Prevents brute force attacks
- **CORS**: Strict origin validation
- **Security headers**: HSTS, CSP, X-Frame-Options, etc.
- **Input validation**: All inputs sanitized and validated

### Infrastructure
- **TLS required**: All connections encrypted in transit
- **Non-root containers**: Reduced attack surface
- **Database isolation**: Not exposed to internet

## Security Best Practices for Self-Hosters

See [Security Policy Documentation](docs/security/SECURITY_POLICY.md) for detailed guidance.

### Quick Checklist:
- [ ] Use HTTPS with valid certificate
- [ ] Set strong, unique JWT_SECRET (32+ characters)
- [ ] Set strong database password
- [ ] Enable database encryption at rest
- [ ] Configure firewall (only expose port 443)
- [ ] Set up automated backups
- [ ] Keep software updated
- [ ] Monitor access logs

## Third-Party Security

### Plaid Integration
- Read-only access to financial data
- Access tokens encrypted with user's key
- Tokens never logged or exposed
- User can disconnect at any time

## Compliance

OwnU is designed to help users comply with:
- **GDPR**: Data export, deletion, consent management
- **CCPA**: Similar rights for California residents

## Security Scanning

This project uses automated security scanning:
- **gosec**: Go static analysis
- **Trivy**: Dependency and container scanning
- **CodeQL**: Advanced code analysis
- **Gitleaks**: Secret detection
- **OWASP ZAP**: Dynamic application testing

## Supply Chain Security

All builds include full provenance tracking:

### Build Provenance (Chalk)
- [Chalk](https://chalkproject.io/) marks binaries with build metadata
- Captures: commit SHA, build timestamp, builder info, SBOM
- Metadata extracted with `chalk extract <binary>`

### Container Signing (Cosign)
- Images signed with Sigstore keyless signing (OIDC)
- Signature verifiable against GitHub Actions identity
- Verify with:
  ```bash
  cosign verify \
    --certificate-identity-regexp "https://github.com/ownu/ownu/*" \
    --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
    ghcr.io/ownu/ownu/backend:latest
  ```

### SBOM (Software Bill of Materials)
- Generated in SPDX and CycloneDX formats
- Attached as attestations to container images
- Available as build artifacts in GitHub Actions

## Bug Bounty

We appreciate security researchers who help keep OwnU secure. While we don't have a formal bug bounty program, we will:
- Credit researchers in release notes (with permission)
- Provide references for security work
- Consider bounties for critical vulnerabilities

## Contact

For security concerns: Open a GitHub Security Advisory or contact the maintainers directly.

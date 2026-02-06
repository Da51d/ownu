# OwnU Security Policy

**Version:** 1.0
**Last Updated:** 2024-01
**Review Frequency:** Annually or after significant changes

## 1. Overview

OwnU is a privacy-first, self-hosted personal finance application. This security policy establishes the controls and practices that protect user data. These policies are designed for individual self-hosters while meeting the requirements of third-party integrations like Plaid.

## 2. Governance and Risk Management

### 2.1 Security Responsibility

For self-hosted deployments, the instance operator is responsible for:
- Keeping the application updated
- Securing the hosting environment
- Managing access credentials
- Monitoring security advisories
- Performing regular backups

### 2.2 Risk Assessment

OwnU conducts ongoing risk assessment through:
- **Automated dependency scanning** - Daily vulnerability scans via Dependabot/Trivy
- **Static code analysis** - On every pull request via CodeQL/gosec
- **Container image scanning** - On every build via Trivy
- **Penetration testing** - Community-driven via bug bounty program

### 2.3 Security Incident Response

1. **Detection** - Automated monitoring and logging detect anomalies
2. **Containment** - Revoke compromised credentials, isolate affected systems
3. **Eradication** - Identify root cause, patch vulnerabilities
4. **Recovery** - Restore from backup if needed, verify integrity
5. **Lessons Learned** - Document incident, update controls

Report security issues to: security@ownu.app (or via GitHub Security Advisories)

### 2.4 Change Management

All changes to OwnU follow:
- Code review required for all changes
- Automated testing (unit, integration, security)
- Staged rollout (dev → staging → production)
- Rollback capability maintained

## 3. Identity and Access Management

### 3.1 Authentication

OwnU uses **WebAuthn/Passkeys** as the primary authentication method:
- **Phishing-resistant** - Cryptographic challenge-response
- **No passwords** - Eliminates password-based attacks
- **Hardware-bound** - Keys tied to authenticator device
- **Multi-factor by design** - Possession + biometric/PIN

Fallback authentication:
- **Recovery phrase** - BIP-39 mnemonic for account recovery
- Recovery phrases are hashed (Argon2id) before storage

### 3.2 Session Management

- JWT tokens with short expiration (15 minutes access, 7 days refresh)
- Tokens include encrypted Data Encryption Key (DEK)
- Sessions invalidated on logout
- Automatic session timeout after inactivity

### 3.3 Authorization

- Row-level security - Users can only access their own data
- All API endpoints require authentication (except health check)
- No shared accounts or delegation supported

### 3.4 Credential Storage

- **No plaintext secrets** - All sensitive data encrypted at rest
- **Plaid access tokens** - Encrypted with user's DEK (AES-256-GCM)
- **Recovery phrases** - Only hash stored (Argon2id)
- **JWT secrets** - Environment variable, never in code

## 4. Infrastructure and Network Security

### 4.1 Network Architecture

```
[Internet] → [TLS Termination/Reverse Proxy] → [Application] → [Database]
                    (nginx/Caddy)                 (Go)         (PostgreSQL)
```

### 4.2 Transport Security

- **TLS 1.2+ required** for all connections
- **HSTS enabled** with 1-year max-age
- **Certificate management** via Let's Encrypt (automatic renewal)
- **No mixed content** - All resources served over HTTPS

### 4.3 Firewall Rules

Default secure configuration:
- Only ports 80 (redirect), 443 (HTTPS) exposed
- Database not exposed to internet (internal network only)
- SSH access restricted to specific IPs (if enabled)

### 4.4 Container Security

- **Non-root user** - Application runs as unprivileged user
- **Read-only filesystem** - Where possible
- **No privileged containers** - Drop all capabilities
- **Resource limits** - CPU/memory limits enforced
- **Image scanning** - Vulnerabilities checked on build

### 4.5 Database Security

- **Network isolation** - Database only accessible from application
- **Encrypted connections** - TLS for database connections
- **Encrypted at rest** - Enable PostgreSQL encryption (or disk-level)
- **Minimal privileges** - Application user has minimal required permissions
- **No default credentials** - Random passwords generated on setup

## 5. Development and Vulnerability Management

### 5.1 Secure Development Lifecycle

1. **Design** - Threat modeling for new features
2. **Develop** - Follow secure coding guidelines (OWASP)
3. **Review** - Mandatory code review with security checklist
4. **Test** - Automated security testing in CI/CD
5. **Deploy** - Automated, reproducible deployments
6. **Monitor** - Runtime security monitoring

### 5.2 Secure Coding Standards

- Input validation on all user input
- Parameterized queries (no SQL injection)
- Output encoding (no XSS)
- CSRF protection via SameSite cookies
- No secrets in code or logs
- Dependency pinning with integrity checks

### 5.3 Vulnerability Management

| Severity | Response Time | Resolution Time |
|----------|---------------|-----------------|
| Critical | 4 hours       | 24 hours        |
| High     | 24 hours      | 7 days          |
| Medium   | 7 days        | 30 days         |
| Low      | 30 days       | 90 days         |

### 5.4 Automated Security Testing

- **SAST** - gosec, CodeQL on every PR
- **SCA** - Dependabot, Trivy for dependency vulnerabilities
- **DAST** - OWASP ZAP in CI pipeline
- **Container scanning** - Trivy on every image build
- **Secret scanning** - Gitleaks, GitHub secret scanning

### 5.5 Penetration Testing

- Annual penetration test (community-driven for open source)
- Bug bounty program for responsible disclosure
- All findings tracked to resolution

### 5.6 Supply Chain Security

OwnU implements comprehensive supply chain security:

**Build Provenance (Chalk)**
- All builds marked with [Chalk](https://chalkproject.io/) for metadata embedding
- Captures: commit SHA, build timestamp, builder identity, environment
- Provenance extractable from any binary with `chalk extract`

**Software Bill of Materials (SBOM)**
- Generated in SPDX and CycloneDX formats
- Includes all dependencies with versions and licenses
- Attached as attestations to container images

**Container Signing (Cosign)**
- Keyless signing via Sigstore OIDC
- Tied to GitHub Actions workflow identity
- Verifiable chain of custody from source to container

**Verification Commands**
```bash
# Verify container signature
cosign verify \
  --certificate-identity-regexp "https://github.com/ownu/ownu/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/ownu/ownu/backend:latest

# Extract build provenance from binary
chalk extract ./ownu
```

## 6. Data Privacy and Protection

### 6.1 Data Classification

| Data Type | Classification | Protection |
|-----------|----------------|------------|
| Financial transactions | Sensitive | Encrypted (AES-256-GCM) |
| Account names/numbers | Sensitive | Encrypted (AES-256-GCM) |
| Plaid access tokens | Secret | Encrypted (AES-256-GCM) |
| Username | Internal | Plaintext (for login) |
| Recovery phrase | Secret | Hashed (Argon2id) |

### 6.2 Encryption

- **Algorithm** - AES-256-GCM for symmetric encryption
- **Key Derivation** - Argon2id for password/phrase-based keys
- **Key Hierarchy**:
  - KEK (Key Encryption Key) - Derived from user authentication
  - DEK (Data Encryption Key) - Random per-user, encrypts data
  - DEK encrypted with KEK, stored in database

### 6.3 Data Minimization

- Only collect data necessary for functionality
- No analytics or telemetry (unless explicitly enabled)
- No third-party trackers
- Financial data never leaves user's infrastructure (except to Plaid for sync)

### 6.4 Data Retention

- Users control their own data retention
- Data deleted immediately on user request
- Plaid connection data removed when disconnected
- Audit logs retained for 90 days (configurable)

### 6.5 Data Subject Rights (GDPR/CCPA)

- **Access** - Export all data via API/UI
- **Rectification** - Edit any personal data
- **Erasure** - Delete account and all data
- **Portability** - Export in standard formats (JSON, CSV)

### 6.6 Third-Party Data Sharing

- **Plaid** - Read-only access to bank transactions (user-initiated)
- No other third-party data sharing
- No data sold or used for advertising

## 7. Logging and Monitoring

### 7.1 Audit Logging

Events logged:
- Authentication attempts (success/failure)
- Authorization failures
- Data access (read/write/delete)
- Configuration changes
- API errors

Log format: Structured JSON with timestamp, user ID, action, resource, outcome

### 7.2 Log Protection

- Logs do not contain sensitive data (passwords, tokens, PII)
- Log integrity protected (append-only, checksums)
- Log retention: 90 days (configurable)
- Logs accessible only to instance operator

### 7.3 Alerting

Automated alerts for:
- Multiple failed authentication attempts
- Unusual access patterns
- Security scan findings
- Certificate expiration (30 days warning)

## 8. Business Continuity

### 8.1 Backup

- Automated daily backups (configurable)
- Backups encrypted at rest
- Backup integrity verified
- Restore tested quarterly

### 8.2 Disaster Recovery

- RTO (Recovery Time Objective): 4 hours
- RPO (Recovery Point Objective): 24 hours
- Recovery procedure documented and tested

## 9. Compliance

### 9.1 Applicable Regulations

- GDPR (if EU users)
- CCPA (if California users)
- PCI-DSS: Not applicable (no card data stored)

### 9.2 Compliance Automation

- Data subject request handling (export, delete)
- Consent management
- Privacy policy generator

## 10. Policy Review

This policy is reviewed:
- Annually
- After security incidents
- After significant architecture changes
- When regulations change

---

## Appendix A: Security Checklist for Self-Hosters

- [ ] Use HTTPS with valid certificate
- [ ] Set strong JWT_SECRET (minimum 32 random characters)
- [ ] Set strong database password
- [ ] Enable database encryption at rest
- [ ] Configure firewall (only expose 443)
- [ ] Set up automated backups
- [ ] Enable audit logging
- [ ] Subscribe to security advisories
- [ ] Keep application updated
- [ ] Review access logs periodically

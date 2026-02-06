# Plaid Security Questionnaire - OwnU Responses

This document provides responses to common Plaid security questionnaire questions for OwnU self-hosted deployments.

## 1. Governance and Risk Management

### 1.1 Do you have a documented security policy?
**Yes.** See [SECURITY_POLICY.md](./SECURITY_POLICY.md)

### 1.2 How often is your security policy reviewed?
**Annually**, or after significant changes to the application architecture.

### 1.3 Do you have an incident response plan?
**Yes.** Documented in Section 2.3 of the Security Policy. Steps include:
1. Detection via automated monitoring
2. Containment by revoking credentials
3. Eradication by patching vulnerabilities
4. Recovery from backups
5. Post-incident review

### 1.4 Do you conduct risk assessments?
**Yes.** Continuous automated scanning plus periodic manual review:
- Daily dependency vulnerability scanning (Dependabot, Trivy)
- Static code analysis on every pull request (gosec, CodeQL)
- Container image scanning on every build

### 1.5 Do you have a designated security contact?
**Yes.** Security issues can be reported via:
- GitHub Security Advisories
- Email (configurable per deployment)

---

## 2. Identity and Access Management

### 2.1 What authentication methods do you use?
**WebAuthn/Passkeys** - FIDO2 compliant, phishing-resistant authentication:
- No passwords stored or transmitted
- Hardware-bound credentials (authenticator device)
- Supports biometric verification
- Falls back to recovery phrase for account recovery

### 2.2 Do you support multi-factor authentication?
**Yes, inherently.** WebAuthn is multi-factor by design:
- Something you have (authenticator device)
- Something you are (biometric) or know (PIN)

### 2.3 How are sessions managed?
- JWT tokens with short expiration (15-minute access tokens)
- Refresh tokens with 7-day expiration
- Automatic session timeout on inactivity
- Sessions invalidated on logout

### 2.4 How is authorization implemented?
- Row-level security: Users can only access their own data
- All API endpoints require authentication (except health check)
- Authorization checked on every request

### 2.5 How are credentials stored?
- **Plaid access tokens**: Encrypted with AES-256-GCM using user's personal DEK
- **Recovery phrases**: Only Argon2id hash stored
- **No plaintext secrets** in database or logs

---

## 3. Infrastructure and Network Security

### 3.1 Is data encrypted in transit?
**Yes.** TLS 1.2+ required for all connections:
- HTTPS enforced with HSTS headers
- Automatic certificate renewal via Let's Encrypt
- Database connections encrypted

### 3.2 Is data encrypted at rest?
**Yes, at the application level:**
- All financial data encrypted with AES-256-GCM
- Per-user encryption keys (DEK)
- DEKs encrypted with user-derived keys (KEK)
- Optional: PostgreSQL encryption at rest

### 3.3 What firewall rules are in place?
Default secure configuration:
- Only ports 80 (redirect to HTTPS) and 443 exposed
- PostgreSQL not exposed to internet
- Docker containers run with minimal privileges

### 3.4 How are containers secured?
- Non-root user in containers
- No privileged containers
- CPU/memory limits enforced
- Container images scanned for vulnerabilities

### 3.5 Is the database exposed to the internet?
**No.** PostgreSQL is only accessible from the application container via internal network.

---

## 4. Development and Vulnerability Management

### 4.1 Do you follow secure coding practices?
**Yes.** Based on OWASP guidelines:
- Input validation on all user input
- Parameterized queries (SQL injection prevention)
- Output encoding (XSS prevention)
- CSRF protection via SameSite cookies
- No secrets in code or logs

### 4.2 What security testing do you perform?
Automated security testing in CI/CD:
| Type | Tool | Frequency |
|------|------|-----------|
| SAST | gosec, CodeQL | Every PR |
| SCA | Dependabot, Trivy | Daily |
| Container Scan | Trivy | Every build |
| Secret Scan | Gitleaks | Every PR |
| DAST | OWASP ZAP | On main branch |

### 4.3 How quickly do you patch vulnerabilities?
| Severity | Response | Resolution |
|----------|----------|------------|
| Critical | 4 hours | 24 hours |
| High | 24 hours | 7 days |
| Medium | 7 days | 30 days |
| Low | 30 days | 90 days |

### 4.4 Do you conduct penetration testing?
**Yes.**
- Community-driven for open source
- Bug bounty program for responsible disclosure
- All findings tracked to resolution

### 4.5 How are dependencies managed?
- Dependencies pinned with integrity checks
- Automated vulnerability scanning
- Automated pull requests for updates (Dependabot)
- Manual review before merging

### 4.6 Do you have supply chain security controls?
**Yes.** Full build provenance and verification:

| Control | Tool | Description |
|---------|------|-------------|
| Build Provenance | Chalk | Embeds metadata (commit, timestamp, builder) in binaries |
| SBOM Generation | Syft/Chalk | SPDX and CycloneDX formats |
| Container Signing | Cosign | Keyless signing via Sigstore OIDC |
| Attestations | Cosign | SBOM attached to container images |

All builds are reproducible and verifiable. Container signatures can be verified with:
```bash
cosign verify --certificate-identity-regexp "https://github.com/ownu/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/ownu/ownu/backend:latest
```

---

## 5. Privacy and Data Protection

### 5.1 What data do you collect?
Minimal data collection:
| Data | Purpose | Storage |
|------|---------|---------|
| Username | Login identifier | Plaintext |
| Financial transactions | Core functionality | Encrypted |
| Account information | Core functionality | Encrypted |
| Plaid tokens | Bank connectivity | Encrypted |

### 5.2 How long is data retained?
- User controls retention (indefinite by default)
- Data deleted immediately on user request
- Audit logs retained 90 days (configurable)

### 5.3 Can users export their data?
**Yes.** GDPR Article 20 compliance:
- Full data export via API (`GET /api/v1/privacy/export`)
- Export formats: JSON, CSV
- Includes all user data, transactions, accounts

### 5.4 Can users delete their data?
**Yes.** GDPR Article 17 compliance:
- Account deletion via API (`DELETE /api/v1/privacy/account`)
- All user data permanently deleted
- Plaid connections removed
- Cannot be undone

### 5.5 Is data shared with third parties?
**Only Plaid** (user-initiated):
- Read-only access to bank data
- User explicitly connects each bank
- User can disconnect at any time
- No other third-party sharing

### 5.6 Do you use analytics or tracking?
**No.**
- No analytics or telemetry by default
- No third-party trackers
- User data never leaves their infrastructure (except Plaid sync)

---

## 6. Business Continuity

### 6.1 Do you perform backups?
**Yes.** Automated backup system:
- Daily encrypted backups (configurable)
- Retention: 30 days (configurable)
- Integrity verification with checksums
- Documented restore procedures

### 6.2 What is your disaster recovery capability?
- RTO: 4 hours
- RPO: 24 hours
- Recovery procedures documented and tested

---

## 7. Compliance

### 7.1 What regulations do you comply with?
- **GDPR**: Data export, deletion, consent management
- **CCPA**: Similar rights for California residents
- **PCI-DSS**: Not applicable (no direct card data storage)

### 7.2 Do you have audit logging?
**Yes.** Events logged include:
- Authentication attempts (success/failure)
- Authorization failures
- Data access (read/write/delete)
- Configuration changes
- API errors

Logs are:
- Structured JSON format
- No sensitive data (passwords, tokens, PII)
- Retained 90 days
- Accessible only to instance operator

---

## 8. Technical Implementation Details

### 8.1 Technology Stack
- **Backend**: Go 1.23+, Echo framework
- **Database**: PostgreSQL 16+
- **Frontend**: React, TypeScript
- **Authentication**: WebAuthn (go-webauthn library)
- **Encryption**: AES-256-GCM, Argon2id
- **Containerization**: Docker, Docker Compose
- **Supply Chain**: Chalk (provenance), Cosign (signing), Syft (SBOM)

### 8.2 Security Headers
All responses include:
- Strict-Transport-Security (HSTS)
- Content-Security-Policy (CSP)
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- Referrer-Policy: strict-origin-when-cross-origin
- Permissions-Policy

### 8.3 Rate Limiting
- 100 requests per minute per IP (configurable)
- Automatic blocking of excessive requests

---

## 9. Verification

### 9.1 How can these controls be verified?
1. **Run security check**: `./scripts/security-check.sh`
2. **Review CI/CD logs**: GitHub Actions security scanning
3. **Inspect code**: Open source, fully auditable
4. **Test APIs**: Security headers, rate limiting visible

### 9.2 Open Source Transparency
OwnU is fully open source (AGPL-3.0), enabling:
- Independent security audits
- Community vulnerability reporting
- Transparent development practices

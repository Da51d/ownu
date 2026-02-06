# OwnU

Privacy-first, self-hosted personal finance tracker.

## Features

- **Passwordless Authentication**: WebAuthn/Passkeys for secure, phishing-resistant login
- **End-to-End Privacy**: Row-level encryption ensures your financial data stays private
- **Self-Hosted**: Run on your own infrastructure, own your data
- **Bank Sync via Plaid**: Automatic transaction import from 12,000+ financial institutions
- **Import Support**: CSV and OFX/QFX file import from any bank
- **Spending Insights**: Categorize transactions and track spending patterns
- **Supply Chain Security**: Signed container images with SBOM and build provenance

## Quick Start

### Prerequisites

- Docker and Docker Compose
- A WebAuthn-compatible browser (Chrome 116+, Safari 17.4+, Firefox)

### Running with Docker

```bash
# Clone the repository
git clone https://github.com/yourusername/ownu.git
cd ownu

# Copy and configure environment
cp .env.example .env
# Edit .env with your settings

# Start all services
docker compose up -d

# Open http://localhost:5173 in your browser
```

### Development Setup

**Backend (Go)**:
```bash
cd backend
go mod download
go run ./cmd/server
```

**Frontend (React)**:
```bash
cd frontend
npm install
npm run dev
```

## Security

OwnU uses a multi-layered security approach:

1. **WebAuthn Authentication**: Passwordless login using device biometrics or security keys
2. **PRF Key Derivation**: Your passkey derives a unique encryption key
3. **AES-256-GCM Encryption**: All sensitive data encrypted before storage
4. **Recovery Phrase**: BIP-39 style backup for account recovery

Your encryption keys never leave your control. Even with full database access, your financial data remains encrypted.

## Plaid Integration

OwnU supports automatic bank sync via [Plaid](https://plaid.com/):

- **Read-only access**: Only transaction data is imported
- **Encrypted tokens**: Plaid access tokens encrypted with your personal key
- **User-controlled**: Connect and disconnect banks at any time
- **12,000+ institutions**: Support for most US/Canadian banks

To enable Plaid, add your API credentials to `.env`. See `.env.example` for details.

## Supply Chain Security

Container images are built with full provenance tracking:

- **Chalk**: Embeds build metadata and generates SBOM during build
- **Cosign**: Keyless signing with Sigstore for image verification
- **SBOM**: Software Bill of Materials in SPDX and CycloneDX formats
- **Attestations**: SBOM attestations attached to container images

Verify an image:
```bash
cosign verify \
  --certificate-identity-regexp "https://github.com/ownu/ownu/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/ownu/ownu/backend:latest
```

## Tech Stack

- **Backend**: Go 1.23 + Echo framework
- **Database**: PostgreSQL 16
- **Frontend**: React + TypeScript + Vite
- **Containerization**: Docker with multi-stage builds
- **Supply Chain**: Chalk (build provenance), Cosign (image signing), Syft (SBOM)

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See [LICENSE](LICENSE) for details.

This means:
- You can use, modify, and distribute this software
- If you modify and deploy it (even as a network service), you must share your changes
- Any derivative work must also be AGPL-3.0 licensed

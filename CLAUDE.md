# OwnU Development Guide

## Project Overview

Privacy-first, self-hosted personal finance tracker. Helps users break the paycheck-to-paycheck cycle.

## Tech Stack

- **Backend**: Go 1.23+ with Echo framework
- **Database**: PostgreSQL 16
- **Frontend**: React 18 + TypeScript + Vite
- **Auth**: WebAuthn/Passkeys (passwordless)
- **Encryption**: AES-256-GCM with Argon2id key derivation
- **Bank Sync**: Plaid API integration
- **Supply Chain**: Chalk (provenance), Cosign (signing), Syft (SBOM)

## Development Commands

### Backend
```bash
cd backend
go run ./cmd/server          # Run server
go test ./...                # Run all tests
go build -o ownu ./cmd/server # Build binary
```

### Frontend
```bash
cd frontend
npm install                  # Install dependencies
npm run dev                  # Development server
npm run build                # Production build
npm test                     # Run tests
```

### Docker
```bash
docker compose up -d         # Start all services
docker compose down          # Stop all services
docker compose logs -f       # View logs
```

### SSL Certificates

**Local Development (self-signed):**
```bash
./scripts/generate-self-signed-cert.sh ./certs localhost
docker compose up -d
# Access at https://localhost (accept browser warning)
```

**Production (Let's Encrypt):**
```bash
# 1. Update DOMAIN in docker-compose.production.yml
# 2. Point your domain's DNS to your server
# 3. Run the initialization script:
./scripts/init-letsencrypt.sh your-domain.com admin@your-domain.com

# Start with production config:
docker compose -f docker-compose.yml -f docker-compose.production.yml up -d
```

## Architecture

### Security Model

1. **WebAuthn PRF Extension**: Derives a Key Encryption Key (KEK) from the user's passkey
2. **Data Encryption Key (DEK)**: Random 256-bit key per user, encrypted with KEK
3. **Row-Level Encryption**: Sensitive fields encrypted with DEK before storage
4. **Server-Side Decryption**: DEK transmitted via JWT, held in memory only during request processing

### Database

All sensitive user data is encrypted at rest. Unencrypted fields are limited to:
- User ID and username
- Transaction dates (for query filtering)
- Foreign key relationships

### API Structure

All endpoints under `/api/v1/`:
- `/auth/*` - WebAuthn registration and login
- `/accounts/*` - Financial account CRUD
- `/transactions/*` - Transaction CRUD
- `/categories/*` - Category management
- `/import/*` - CSV/OFX import
- `/reports/*` - Spending and cashflow reports
- `/plaid/*` - Plaid bank integration (link, sync, manage)
- `/privacy/*` - GDPR/CCPA compliance (export, delete)

## Code Conventions

- Go: Follow standard Go conventions, use `gofmt`
- TypeScript: Use strict mode, prefer interfaces over types
- API responses: Always return JSON with consistent error format
- Encryption: Never log or persist decrypted sensitive data

## Environment Variables

See `.env.example` for required configuration.

## Testing

- Backend: Table-driven tests, mock database for unit tests
- Frontend: React Testing Library for component tests
- Integration: Docker Compose for full-stack testing

## Security Scanning

CI/CD pipeline includes:
- **gosec** - Go static analysis
- **Trivy** - Dependency and container scanning
- **CodeQL** - GitHub advanced security
- **Gitleaks** - Secret detection
- **OWASP ZAP** - Dynamic application testing

## Supply Chain Security

Builds use [Chalk](https://chalkproject.io/) for provenance:
- Binary marked with `chalk insert` during build
- Docker builds wrapped with `chalk docker build`
- Generates SBOM in SPDX and CycloneDX formats
- Container images signed with Cosign (keyless/OIDC)

Extract chalk metadata from a binary:
```bash
chalk extract ./ownu
```

Verify container signature:
```bash
cosign verify --certificate-identity-regexp "https://github.com/ownu/*" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  ghcr.io/ownu/ownu/backend:latest
```

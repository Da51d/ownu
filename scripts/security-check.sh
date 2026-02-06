#!/bin/bash
# OwnU Security Configuration Checker
# Validates that security best practices are followed
#
# Usage: ./security-check.sh

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASSED=0
FAILED=0
WARNINGS=0

pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
    ((WARNINGS++))
}

info() {
    echo -e "[INFO] $1"
}

echo "=== OwnU Security Configuration Check ==="
echo ""

# 1. Check JWT_SECRET
info "Checking JWT configuration..."
if [ -z "${JWT_SECRET:-}" ]; then
    fail "JWT_SECRET is not set"
elif [ ${#JWT_SECRET} -lt 32 ]; then
    fail "JWT_SECRET is too short (minimum 32 characters)"
elif [ "$JWT_SECRET" = "change_me_in_production" ]; then
    fail "JWT_SECRET is using default value"
else
    pass "JWT_SECRET is configured properly (${#JWT_SECRET} characters)"
fi

# 2. Check DATABASE_URL
info "Checking database configuration..."
if [ -z "${DATABASE_URL:-}" ]; then
    fail "DATABASE_URL is not set"
else
    if echo "$DATABASE_URL" | grep -q "sslmode=disable"; then
        warn "Database SSL is disabled (consider enabling for production)"
    else
        pass "Database URL is configured"
    fi

    if echo "$DATABASE_URL" | grep -qE "(localhost|127\.0\.0\.1)"; then
        pass "Database is on localhost (not exposed to internet)"
    else
        warn "Database appears to be on remote host - ensure network security"
    fi
fi

# 3. Check WebAuthn configuration
info "Checking WebAuthn configuration..."
if [ -z "${WEBAUTHN_RP_ID:-}" ]; then
    fail "WEBAUTHN_RP_ID is not set"
else
    pass "WEBAUTHN_RP_ID is set: $WEBAUTHN_RP_ID"
fi

if [ -z "${WEBAUTHN_RP_ORIGIN:-}" ]; then
    fail "WEBAUTHN_RP_ORIGIN is not set"
elif echo "$WEBAUTHN_RP_ORIGIN" | grep -q "^http://"; then
    warn "WEBAUTHN_RP_ORIGIN uses HTTP (should use HTTPS in production)"
else
    pass "WEBAUTHN_RP_ORIGIN uses HTTPS"
fi

# 4. Check Plaid configuration (optional)
info "Checking Plaid configuration..."
if [ -n "${PLAID_CLIENT_ID:-}" ] && [ -n "${PLAID_SECRET:-}" ]; then
    pass "Plaid credentials are configured"

    if [ "${PLAID_ENV:-sandbox}" = "production" ]; then
        info "Plaid is in PRODUCTION mode"
    else
        info "Plaid is in ${PLAID_ENV:-sandbox} mode"
    fi
else
    info "Plaid is not configured (optional)"
fi

# 5. Check file permissions
info "Checking file permissions..."
if [ -f ".env" ]; then
    PERMS=$(stat -c "%a" .env 2>/dev/null || stat -f "%OLp" .env 2>/dev/null)
    if [ "$PERMS" = "600" ] || [ "$PERMS" = "400" ]; then
        pass ".env file has restricted permissions ($PERMS)"
    else
        warn ".env file permissions are too open ($PERMS) - should be 600"
    fi
fi

# 6. Check for common security files
info "Checking security documentation..."
if [ -f "SECURITY.md" ] || [ -f "docs/security/SECURITY_POLICY.md" ]; then
    pass "Security policy documentation exists"
else
    warn "No security policy documentation found"
fi

# 7. Check Docker configuration
info "Checking Docker configuration..."
if [ -f "docker-compose.yml" ]; then
    if grep -q "privileged: true" docker-compose.yml; then
        fail "Docker Compose uses privileged containers"
    else
        pass "No privileged containers in Docker Compose"
    fi

    if grep -q "network_mode: host" docker-compose.yml; then
        warn "Docker Compose uses host network mode"
    fi
fi

if [ -f "backend/Dockerfile" ]; then
    if grep -q "USER" backend/Dockerfile; then
        pass "Backend Dockerfile uses non-root user"
    else
        warn "Backend Dockerfile may run as root"
    fi
fi

# 8. Check for exposed ports
info "Checking network exposure..."
if command -v ss &> /dev/null; then
    # Check if PostgreSQL is listening on public interfaces
    if ss -tlnp 2>/dev/null | grep -q "0.0.0.0:5432"; then
        fail "PostgreSQL is listening on all interfaces (0.0.0.0)"
    elif ss -tlnp 2>/dev/null | grep -q ":5432"; then
        pass "PostgreSQL is not exposed publicly"
    fi
fi

# 9. Check SSL/TLS certificate
info "Checking TLS configuration..."
if [ -d "/etc/letsencrypt/live" ]; then
    # Check certificate expiry
    for domain_dir in /etc/letsencrypt/live/*/; do
        if [ -f "${domain_dir}cert.pem" ]; then
            EXPIRY=$(openssl x509 -enddate -noout -in "${domain_dir}cert.pem" 2>/dev/null | cut -d= -f2)
            EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$EXPIRY" +%s 2>/dev/null)
            NOW_EPOCH=$(date +%s)
            DAYS_LEFT=$(( (EXPIRY_EPOCH - NOW_EPOCH) / 86400 ))

            if [ $DAYS_LEFT -lt 7 ]; then
                fail "TLS certificate expires in $DAYS_LEFT days"
            elif [ $DAYS_LEFT -lt 30 ]; then
                warn "TLS certificate expires in $DAYS_LEFT days"
            else
                pass "TLS certificate valid for $DAYS_LEFT days"
            fi
        fi
    done
else
    info "Let's Encrypt not found (may be using different certificate provider)"
fi

# 10. Check backup configuration
info "Checking backup configuration..."
if [ -d "/var/backups/ownu" ]; then
    BACKUP_COUNT=$(find /var/backups/ownu -name "ownu_backup_*.dump*" 2>/dev/null | wc -l)
    if [ $BACKUP_COUNT -gt 0 ]; then
        LATEST_BACKUP=$(find /var/backups/ownu -name "ownu_backup_*.dump*" -printf '%T@ %p\n' 2>/dev/null | sort -n | tail -1 | cut -d' ' -f2)
        BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_BACKUP" 2>/dev/null || stat -f %m "$LATEST_BACKUP" 2>/dev/null)) / 86400 ))

        if [ $BACKUP_AGE -gt 7 ]; then
            warn "Latest backup is $BACKUP_AGE days old"
        else
            pass "Backups exist, latest is $BACKUP_AGE days old"
        fi
    else
        warn "No backups found in /var/backups/ownu"
    fi
else
    info "Backup directory not found (run backup.sh to create)"
fi

# Summary
echo ""
echo "=== Security Check Summary ==="
echo -e "Passed:   ${GREEN}$PASSED${NC}"
echo -e "Failed:   ${RED}$FAILED${NC}"
echo -e "Warnings: ${YELLOW}$WARNINGS${NC}"
echo ""

if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Security check FAILED - please address the issues above${NC}"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}Security check passed with warnings${NC}"
    exit 0
else
    echo -e "${GREEN}Security check PASSED${NC}"
    exit 0
fi

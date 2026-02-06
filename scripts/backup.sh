#!/bin/bash
# OwnU Backup Script
# Creates encrypted backups of the PostgreSQL database
#
# Usage: ./backup.sh [backup_dir]
#
# Environment variables:
#   DATABASE_URL - PostgreSQL connection string
#   BACKUP_ENCRYPTION_KEY - Key for encrypting backups (optional)

set -euo pipefail

# Configuration
BACKUP_DIR="${1:-/var/backups/ownu}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="ownu_backup_${TIMESTAMP}.sql"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check required tools
check_requirements() {
    if ! command -v pg_dump &> /dev/null; then
        log_error "pg_dump is not installed"
        exit 1
    fi

    if [ -n "${BACKUP_ENCRYPTION_KEY:-}" ] && ! command -v openssl &> /dev/null; then
        log_error "openssl is not installed (required for encryption)"
        exit 1
    fi
}

# Parse DATABASE_URL
parse_database_url() {
    if [ -z "${DATABASE_URL:-}" ]; then
        log_error "DATABASE_URL environment variable is not set"
        exit 1
    fi

    # Export for pg_dump
    export PGPASSWORD=$(echo "$DATABASE_URL" | sed -n 's|.*://[^:]*:\([^@]*\)@.*|\1|p')
    DB_HOST=$(echo "$DATABASE_URL" | sed -n 's|.*@\([^:/]*\).*|\1|p')
    DB_PORT=$(echo "$DATABASE_URL" | sed -n 's|.*:\([0-9]*\)/.*|\1|p')
    DB_NAME=$(echo "$DATABASE_URL" | sed -n 's|.*/\([^?]*\).*|\1|p')
    DB_USER=$(echo "$DATABASE_URL" | sed -n 's|.*://\([^:]*\):.*|\1|p')

    # Default port if not specified
    DB_PORT="${DB_PORT:-5432}"
}

# Create backup directory
create_backup_dir() {
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        chmod 700 "$BACKUP_DIR"
        log_info "Created backup directory: $BACKUP_DIR"
    fi
}

# Perform backup
perform_backup() {
    log_info "Starting backup of database: $DB_NAME"

    # Perform the dump
    pg_dump \
        -h "$DB_HOST" \
        -p "$DB_PORT" \
        -U "$DB_USER" \
        -d "$DB_NAME" \
        --format=custom \
        --no-owner \
        --no-acl \
        -f "${BACKUP_DIR}/${BACKUP_FILE}.dump"

    if [ $? -ne 0 ]; then
        log_error "Database backup failed"
        exit 1
    fi

    log_info "Database dump created: ${BACKUP_FILE}.dump"

    # Encrypt if key is provided
    if [ -n "${BACKUP_ENCRYPTION_KEY:-}" ]; then
        log_info "Encrypting backup..."
        openssl enc -aes-256-cbc -salt -pbkdf2 \
            -in "${BACKUP_DIR}/${BACKUP_FILE}.dump" \
            -out "${BACKUP_DIR}/${BACKUP_FILE}.dump.enc" \
            -pass pass:"$BACKUP_ENCRYPTION_KEY"

        if [ $? -eq 0 ]; then
            rm "${BACKUP_DIR}/${BACKUP_FILE}.dump"
            BACKUP_FILE="${BACKUP_FILE}.dump.enc"
            log_info "Backup encrypted: $BACKUP_FILE"
        else
            log_error "Encryption failed, keeping unencrypted backup"
        fi
    else
        BACKUP_FILE="${BACKUP_FILE}.dump"
        log_warn "Backup is not encrypted (set BACKUP_ENCRYPTION_KEY to enable)"
    fi

    # Create checksum
    sha256sum "${BACKUP_DIR}/${BACKUP_FILE}" > "${BACKUP_DIR}/${BACKUP_FILE}.sha256"
    log_info "Checksum created: ${BACKUP_FILE}.sha256"
}

# Clean old backups
cleanup_old_backups() {
    log_info "Cleaning backups older than $RETENTION_DAYS days..."

    find "$BACKUP_DIR" -name "ownu_backup_*.dump*" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
    find "$BACKUP_DIR" -name "ownu_backup_*.sha256" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true

    # Count remaining backups
    BACKUP_COUNT=$(find "$BACKUP_DIR" -name "ownu_backup_*.dump*" | wc -l)
    log_info "Remaining backups: $BACKUP_COUNT"
}

# Verify backup
verify_backup() {
    log_info "Verifying backup integrity..."

    cd "$BACKUP_DIR"
    if sha256sum -c "${BACKUP_FILE}.sha256" &>/dev/null; then
        log_info "Backup verification: PASSED"
    else
        log_error "Backup verification: FAILED"
        exit 1
    fi
}

# Main
main() {
    log_info "=== OwnU Backup Script ==="

    check_requirements
    parse_database_url
    create_backup_dir
    perform_backup
    verify_backup
    cleanup_old_backups

    log_info "=== Backup completed successfully ==="
    log_info "Backup location: ${BACKUP_DIR}/${BACKUP_FILE}"
}

main "$@"

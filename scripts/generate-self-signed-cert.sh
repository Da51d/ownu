#!/bin/bash
# Generate self-signed SSL certificates for local development

set -e

CERT_DIR="${1:-./certs}"
DOMAIN="${2:-localhost}"

mkdir -p "$CERT_DIR"

# Generate private key and self-signed certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$CERT_DIR/privkey.pem" \
    -out "$CERT_DIR/fullchain.pem" \
    -subj "/C=US/ST=Local/L=Local/O=OwnU/CN=$DOMAIN" \
    -addext "subjectAltName=DNS:$DOMAIN,DNS:www.$DOMAIN,IP:127.0.0.1"

# Generate DH parameters for enhanced security
openssl dhparam -out "$CERT_DIR/dhparam.pem" 2048

echo "Self-signed certificates generated in $CERT_DIR"
echo "  - fullchain.pem (certificate)"
echo "  - privkey.pem (private key)"
echo "  - dhparam.pem (DH parameters)"

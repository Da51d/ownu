#!/bin/bash
# Initialize Let's Encrypt certificates using Certbot
# Run this script when you have a real domain pointed to your server

set -e

DOMAIN="${1:-}"
EMAIL="${2:-}"
STAGING="${3:-0}"

if [ -z "$DOMAIN" ] || [ -z "$EMAIL" ]; then
    echo "Usage: $0 <domain> <email> [staging]"
    echo "  domain:  Your domain name (e.g., ownu.example.com)"
    echo "  email:   Email for Let's Encrypt notifications"
    echo "  staging: Set to 1 to use Let's Encrypt staging (for testing)"
    exit 1
fi

DATA_PATH="./certbot"
RSA_KEY_SIZE=4096

# Check if certificates already exist
if [ -d "$DATA_PATH/conf/live/$DOMAIN" ]; then
    echo "Certificates already exist for $DOMAIN"
    read -p "Replace existing certificates? (y/N) " decision
    if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
        exit 0
    fi
fi

# Create required directories
mkdir -p "$DATA_PATH/conf/live/$DOMAIN"
mkdir -p "$DATA_PATH/www"

# Download recommended TLS parameters
if [ ! -e "$DATA_PATH/conf/options-ssl-nginx.conf" ]; then
    echo "Downloading recommended TLS parameters..."
    curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$DATA_PATH/conf/options-ssl-nginx.conf"
fi

if [ ! -e "$DATA_PATH/conf/ssl-dhparams.pem" ]; then
    echo "Downloading DH parameters..."
    curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$DATA_PATH/conf/ssl-dhparams.pem"
fi

# Create dummy certificate for nginx to start
echo "Creating dummy certificate for $DOMAIN..."
openssl req -x509 -nodes -newkey rsa:$RSA_KEY_SIZE -days 1 \
    -keyout "$DATA_PATH/conf/live/$DOMAIN/privkey.pem" \
    -out "$DATA_PATH/conf/live/$DOMAIN/fullchain.pem" \
    -subj "/CN=localhost"

echo "Starting nginx..."
docker compose up -d frontend

echo "Deleting dummy certificate..."
rm -rf "$DATA_PATH/conf/live/$DOMAIN"

echo "Requesting Let's Encrypt certificate for $DOMAIN..."

# Set staging flag if requested
STAGING_ARG=""
if [ "$STAGING" = "1" ]; then
    STAGING_ARG="--staging"
    echo "Using Let's Encrypt staging environment"
fi

docker compose run --rm certbot certonly --webroot \
    --webroot-path=/var/www/certbot \
    $STAGING_ARG \
    --email "$EMAIL" \
    --agree-tos \
    --no-eff-email \
    -d "$DOMAIN"

echo "Reloading nginx..."
docker compose exec frontend nginx -s reload

echo ""
echo "SSL certificate obtained successfully!"
echo "Your site is now available at https://$DOMAIN"

#!/bin/sh
set -e

CERT_DIR="/app/certs"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate self-signed certificate if it doesn't exist
if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "🔐 Generating self-signed TLS certificate..."
    
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$KEY_FILE" \
        -out "$CERT_FILE" \
        -subj "/C=US/ST=Local/L=Local/O=RAG-Assistant/OU=Development/CN=localhost" \
        -addext "subjectAltName=DNS:localhost,DNS:backend,IP:127.0.0.1"
    
    chmod 600 "$KEY_FILE"
    chmod 644 "$CERT_FILE"
    
    echo "TLS certificate generated successfully"
else
    echo "TLS certificate already exists"
fi

# Export cert paths for the application
export TLS_CERT_FILE="$CERT_FILE"
export TLS_KEY_FILE="$KEY_FILE"

# Execute the main application
exec "$@"

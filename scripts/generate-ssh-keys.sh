#!/bin/bash
# Generate SSH key pair for Orchestra test servers
# Usage: ./scripts/generate-ssh-keys.sh

set -e

KEYS_DIR="$(cd "$(dirname "$0")/.." && pwd)/keys"
PRIVATE_KEY="$KEYS_DIR/id_rsa"
PUBLIC_KEY="$KEYS_DIR/id_rsa.pub"

mkdir -p "$KEYS_DIR"

if [ -f "$PRIVATE_KEY" ]; then
    echo "[*] SSH keys already exist at $KEYS_DIR"
    echo "    Private: $PRIVATE_KEY"
    echo "    Public:  $PUBLIC_KEY"
    echo ""
    echo "To regenerate, remove the keys first: rm -rf $KEYS_DIR"
    exit 0
fi

echo "[*] Generating SSH key pair..."
ssh-keygen -t rsa -b 4096 -f "$PRIVATE_KEY" -N "" -C "orchestra-test"

echo ""
echo "[*] Keys created:"
echo "    Private: $PRIVATE_KEY  (use this when registering servers in Orchestra)"
echo "    Public:  $PUBLIC_KEY   (added to test servers automatically)"
echo ""
echo "Next steps:"
echo "  1. Start test servers: docker compose -f docker-compose.yml -f docker-compose.test-servers.yml up -d"
echo "  2. Open Orchestra UI: http://localhost:3000"
echo "  3. Register each test server with:"
echo "     - IP: test-server-1 (or test-server-2, test-server-3)"
echo "     - Port: 22"
echo "     - User: orchestra"
echo "     - Key: paste contents of $PRIVATE_KEY"
echo ""

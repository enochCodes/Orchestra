#!/bin/bash
# setup-ssh-container.sh — Create a standalone SSH container for testing Orchestra
# Usage: ./setup-ssh-container.sh [container_name] [ssh_user] [ssh_password] [ssh_port] [pubkey_file]
#
# For Orchestra test servers, prefer using docker-compose:
#   ./scripts/generate-ssh-keys.sh
#   docker compose -f docker-compose.yml -f docker-compose.test-servers.yml up -d

CONTAINER_NAME=${1:-test-ssh}
SSH_USER=${2:-orchestra}
SSH_PASS=${3:-orchestra}
SSH_PORT=${4:-2222}
PUBKEY_FILE=${5:-""}

# Generate key pair if not provided and keys/ exists
if [ -z "$PUBKEY_FILE" ] && [ -f "keys/id_rsa.pub" ]; then
    PUBKEY_FILE="keys/id_rsa.pub"
    echo "[*] Using existing key: $PUBKEY_FILE"
elif [ -z "$PUBKEY_FILE" ]; then
    echo "[*] No public key provided. Run ./scripts/generate-ssh-keys.sh first for key-based auth."
fi

# Pull lightweight Ubuntu image
docker pull ubuntu:22.04

# Create container with port mapping
docker run -d --name "$CONTAINER_NAME" -p "$SSH_PORT:22" ubuntu:22.04 sleep infinity

# Install SSH server and create user
echo "[*] Installing SSH server in $CONTAINER_NAME..."
docker exec "$CONTAINER_NAME" bash -c "
apt-get update && apt-get install -y openssh-server sudo && \
mkdir -p /var/run/sshd && \
useradd -m -s /bin/bash $SSH_USER && \
echo '$SSH_USER:$SSH_PASS' | chpasswd && \
usermod -aG sudo $SSH_USER
"

# Enable password and key auth
docker exec "$CONTAINER_NAME" bash -c "
sed -i 's/^#*PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config && \
sed -i 's/^#*PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config && \
sed -i 's/^#*PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config
"

# Add public key if provided
if [ -n "$PUBKEY_FILE" ] && [ -f "$PUBKEY_FILE" ]; then
    echo "[*] Adding public key for $SSH_USER..."
    PUBKEY_CONTENT=$(cat "$PUBKEY_FILE")
    docker exec "$CONTAINER_NAME" bash -c "
    mkdir -p /home/$SSH_USER/.ssh && \
    echo '$PUBKEY_CONTENT' >> /home/$SSH_USER/.ssh/authorized_keys && \
    chown -R $SSH_USER:$SSH_USER /home/$SSH_USER/.ssh && \
    chmod 700 /home/$SSH_USER/.ssh && chmod 600 /home/$SSH_USER/.ssh/authorized_keys
    "
fi

# Start SSH server
docker exec -d "$CONTAINER_NAME" /usr/sbin/sshd -D

CONTAINER_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$CONTAINER_NAME")

echo ""
echo "[*] SSH container '$CONTAINER_NAME' is ready!"
echo "    Connect from host:     ssh $SSH_USER@localhost -p $SSH_PORT"
echo "    For Orchestra:         IP=localhost, Port=$SSH_PORT, User=$SSH_USER"
echo "    Private key (if used): keys/id_rsa"
echo ""

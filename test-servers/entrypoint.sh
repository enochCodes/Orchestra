#!/bin/bash
set -e

# Add public key to authorized_keys if provided
if [ -n "$ORCHESTRA_PUBLIC_KEY" ]; then
    mkdir -p /home/orchestra/.ssh
    echo "$ORCHESTRA_PUBLIC_KEY" >> /home/orchestra/.ssh/authorized_keys
    chown -R orchestra:orchestra /home/orchestra/.ssh
    chmod 700 /home/orchestra/.ssh
    chmod 600 /home/orchestra/.ssh/authorized_keys
fi

# If key file is mounted
if [ -f /keys/id_rsa.pub ]; then
    mkdir -p /home/orchestra/.ssh
    cat /keys/id_rsa.pub >> /home/orchestra/.ssh/authorized_keys
    chown -R orchestra:orchestra /home/orchestra/.ssh
    chmod 700 /home/orchestra/.ssh
    chmod 600 /home/orchestra/.ssh/authorized_keys
fi

exec /usr/sbin/sshd -D

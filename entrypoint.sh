#!/bin/sh
set -e

KEY_PATH="${SSH_HOST_KEY:-/.ssh/ssh}"
if [ ! -f "$KEY_PATH" ]; then
    echo "Host key not found at $KEY_PATH, generating..."
    mkdir -p "$(dirname "$KEY_PATH")"
    ssh-keygen -t ed25519 -f "$KEY_PATH" -N '' -C "pratyaysh@$(hostname)" || {
        echo "Failed to generate ed25519, trying rsa..."
        ssh-keygen -t rsa -b 4096 -f "$KEY_PATH" -N ''
    }
    chmod 600 "$KEY_PATH"
fi

# Start Cloudflare Tunnel in the background if TUNNEL_TOKEN is set
if [ -n "$TUNNEL_TOKEN" ]; then
    echo "Starting Cloudflare Tunnel..."
    cloudflared tunnel --no-autoupdate run --token "$TUNNEL_TOKEN" &
fi

exec pratyaysh "$@"

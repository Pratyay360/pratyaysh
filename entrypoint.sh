#!/bin/sh
set -e

# Ensure SSH host key exists (Railway has ephemeral FS without volume).
# Prefer /.ssh/ssh (container path) then fallback.
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
    # Use quic by default, but Railway blocks some UDP - cloudflared auto-falls back to http2.
    # Optionally force http2: --protocol http2
    cloudflared tunnel --no-autoupdate run --token "$TUNNEL_TOKEN" &
fi

# Execute the main application
exec pratyaysh "$@"

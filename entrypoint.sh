#!/bin/sh
set -e

# Start Cloudflare Tunnel in the background if TUNNEL_TOKEN is set
if [ -n "$TUNNEL_TOKEN" ]; then
    echo "Starting Cloudflare Tunnel..."
    cloudflared tunnel --no-autoupdate run --token "$TUNNEL_TOKEN" &
fi

# Execute the main application
exec pratyaysh "$@"

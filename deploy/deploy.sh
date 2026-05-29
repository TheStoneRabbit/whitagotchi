#!/usr/bin/env bash
# Build the server for linux/amd64 and rsync to your VPS, then restart the service.
#
# Usage:
#   HOST=you@yourbox.com ./deploy/deploy.sh
#
# Requirements on the box (one-time setup, see deploy/README.md):
#   - user `whitagotchi` exists, owns /var/lib/whitagotchi
#   - /etc/systemd/system/whitagotchi-server.service installed and enabled
#   - port 8080 open in the firewall

set -euo pipefail

: "${HOST:?set HOST=user@yourbox.com}"
REMOTE_BIN="${REMOTE_BIN:-/usr/local/bin/whitagotchi-server}"

echo "==> building linux/amd64 binary"
(
  cd server
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ../dist/whitagotchi-server-linux-amd64 ./cmd/whitagotchi-server
)

echo "==> uploading to $HOST:$REMOTE_BIN"
scp dist/whitagotchi-server-linux-amd64 "$HOST:/tmp/whitagotchi-server.new"
ssh "$HOST" "sudo install -m 0755 /tmp/whitagotchi-server.new $REMOTE_BIN && rm /tmp/whitagotchi-server.new"

echo "==> restarting service"
ssh "$HOST" "sudo systemctl restart whitagotchi-server && sudo systemctl status --no-pager whitagotchi-server"

echo "==> done"

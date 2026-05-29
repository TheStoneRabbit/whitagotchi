#!/usr/bin/env bash
# Build server + client binaries, ship them to the VPS, and restart the service.
#
# Usage:
#   HOST=root@yourbox.com ./deploy/deploy.sh
#
# Optional env:
#   REMOTE_BIN     path to install server binary (default /usr/local/bin/whitagotchi-server)
#   CLIENT_BIN_DIR remote dir the server serves at /bin/ (default /var/lib/whitagotchi/bin)
#   SKIP_CLIENTS=1 skip cross-compiling client binaries (faster server-only deploys)

set -euo pipefail

: "${HOST:?set HOST=user@yourbox.com}"
REMOTE_BIN="${REMOTE_BIN:-/usr/local/bin/whitagotchi-server}"
CLIENT_BIN_DIR="${CLIENT_BIN_DIR:-/var/lib/whitagotchi/bin}"

echo "==> building server (linux/amd64)"
(
  cd server
  GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ../dist/whitagotchi-server-linux-amd64 ./cmd/whitagotchi-server
)

echo "==> uploading server to $HOST:$REMOTE_BIN"
scp dist/whitagotchi-server-linux-amd64 "$HOST:/tmp/whitagotchi-server.new"
ssh "$HOST" "sudo install -m 0755 /tmp/whitagotchi-server.new $REMOTE_BIN && rm /tmp/whitagotchi-server.new"

if [ "${SKIP_CLIENTS:-0}" != "1" ]; then
  echo "==> building client binaries"
  mkdir -p dist
  (
    cd client
    for target in "darwin amd64" "darwin arm64" "linux amd64"; do
      set -- $target
      os="$1"; arch="$2"
      out="../dist/whitagotchi-${os}-${arch}"
      echo "  - $os/$arch -> $out"
      GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
        go build -ldflags="-s -w" -o "$out" ./cmd/whitagotchi
    done
  )

  echo "==> uploading client binaries to $HOST:$CLIENT_BIN_DIR"
  ssh "$HOST" "sudo install -d -o whitagotchi -g whitagotchi $CLIENT_BIN_DIR"
  for bin in dist/whitagotchi-darwin-amd64 dist/whitagotchi-darwin-arm64 dist/whitagotchi-linux-amd64; do
    name=$(basename "$bin")
    scp "$bin" "$HOST:/tmp/$name"
    ssh "$HOST" "sudo install -m 0755 -o whitagotchi -g whitagotchi /tmp/$name $CLIENT_BIN_DIR/$name && rm /tmp/$name"
  done
fi

echo "==> restarting service"
ssh "$HOST" "sudo systemctl restart whitagotchi-server && sudo systemctl status --no-pager whitagotchi-server | head -10"

echo "==> done"

package api

import (
	"fmt"
	"net/http"
	"strings"
)

const installScript = `#!/usr/bin/env sh
# whitagotchi installer — fetches the right client binary for your OS/arch
# from the server you got this script from, and installs it on your PATH.
#
#   curl -fsSL %[1]s/install | sh

set -eu

BASE="%[1]s"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "unsupported os: $os" >&2; exit 1 ;;
esac

url="$BASE/bin/whitagotchi-${os}-${arch}"

dest_dir="$HOME/.local/bin"
if [ -w /usr/local/bin ]; then dest_dir="/usr/local/bin"; fi
mkdir -p "$dest_dir"
dest="$dest_dir/whitagotchi"

echo "downloading $url"
if ! curl -fsSL -o "$dest.new" "$url"; then
  echo "download failed (is a binary uploaded for ${os}-${arch}?)" >&2
  rm -f "$dest.new"
  exit 1
fi
chmod +x "$dest.new"
mv "$dest.new" "$dest"

echo "installed: $dest"

# Persist the server URL so the client knows where to connect.
config_dir="$HOME/.config/whitagotchi"
if [ "$(uname -s)" = "Darwin" ]; then
  config_dir="$HOME/Library/Application Support/whitagotchi"
fi
mkdir -p "$config_dir"
if [ ! -f "$config_dir/config.json" ]; then
  printf '{"server":"%[1]s"}\n' > "$config_dir/config.json"
  echo "server set to $BASE in $config_dir/config.json"
fi

case ":$PATH:" in
  *":$dest_dir:"*) ;;
  *) echo "warning: $dest_dir is not on your PATH" ;;
esac

echo "next: whitagotchi register <username>"
`

func (s *Server) install(w http.ResponseWriter, r *http.Request) {
	base := s.installer.PublicURL
	if base == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		base = scheme + "://" + r.Host
	}
	w.Header().Set("Content-Type", "text/x-shellscript")
	fmt.Fprintf(w, installScript, base)
}

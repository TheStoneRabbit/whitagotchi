#!/usr/bin/env sh
# whitagotchi installer
#
# Usage:  curl -fsSL https://raw.githubusercontent.com/<owner>/whitagotchi/main/install.sh | sh
#
# Detects OS/arch, downloads the matching binary from the latest GitHub release,
# and drops it in $HOME/.local/bin (or /usr/local/bin if writable).

set -eu

REPO="${WHITAGOTCHI_REPO:-CHANGE_ME/whitagotchi}"
VERSION="${WHITAGOTCHI_VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "unsupported os: $os (use the native binary on windows)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/$REPO/releases/latest/download/whitagotchi-${os}-${arch}"
else
  url="https://github.com/$REPO/releases/download/$VERSION/whitagotchi-${os}-${arch}"
fi

dest_dir="$HOME/.local/bin"
if [ -w /usr/local/bin ]; then
  dest_dir="/usr/local/bin"
fi
mkdir -p "$dest_dir"
dest="$dest_dir/whitagotchi"

echo "downloading $url"
curl -fsSL "$url" -o "$dest"
chmod +x "$dest"
echo "installed: $dest"
echo "make sure $dest_dir is on your PATH"

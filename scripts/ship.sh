#!/usr/bin/env bash
# ship: one command from laptop -> live server + updated local client.
#
# Usage:
#   ./scripts/ship.sh                          # ship whatever's committed locally
#   ./scripts/ship.sh -m "fix the thing"       # commit all changes first, then ship
#   SKIP_CLIENTS=1 ./scripts/ship.sh           # server-only deploy (faster)
#   HOST=root@otherbox ./scripts/ship.sh       # different VPS
#
# Steps:
#   1. (optional) commit current changes with -m
#   2. push to origin
#   3. cross-compile + deploy server and client binaries to the VPS
#   4. reinstall the local client from the deployed server
#   5. nuke any old ~/go/bin/whitagotchi that would shadow the new install

set -euo pipefail

HOST="${HOST:-root@masonlapine.com}"
SERVER_URL="${SERVER_URL:-http://${HOST#*@}:8080}"

# parse -m
commit_msg=""
while getopts "m:" opt; do
  case "$opt" in
    m) commit_msg="$OPTARG" ;;
    *) echo "usage: $0 [-m \"commit message\"]" >&2; exit 2 ;;
  esac
done

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

# 1. commit, if asked
if [ -n "$commit_msg" ]; then
  if [ -z "$(git status --porcelain)" ]; then
    echo "==> nothing to commit; skipping"
  else
    echo "==> committing local changes"
    git add -A
    git commit -m "$commit_msg"
  fi
fi

# Refuse to ship a dirty tree unless -m was used to commit it.
if [ -n "$(git status --porcelain)" ]; then
  echo "error: working tree has uncommitted changes." >&2
  echo "       either commit manually, or rerun with: $0 -m \"your message\"" >&2
  exit 1
fi

# 2. push
echo "==> pushing to origin"
git push

# 3. deploy
echo "==> deploying to $HOST"
HOST="$HOST" "$repo_root/deploy/deploy.sh"

# 4. reinstall local client (skip if we just did a server-only deploy)
if [ "${SKIP_CLIENTS:-0}" != "1" ]; then
  echo "==> reinstalling local client from $SERVER_URL"
  curl -fsSL "$SERVER_URL/install" | sh

  # 5. clear PATH shadows from `go install` runs
  if [ -f "$HOME/go/bin/whitagotchi" ]; then
    echo "==> removing stale $HOME/go/bin/whitagotchi (PATH shadow)"
    rm "$HOME/go/bin/whitagotchi"
  fi
  hash -r 2>/dev/null || true
fi

echo "==> shipped"
which whitagotchi || true

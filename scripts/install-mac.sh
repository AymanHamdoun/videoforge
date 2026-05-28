#!/usr/bin/env bash
#
# Builds VideoForge and installs it to /Applications so it shows up in Spotlight
# and Launchpad (and stays self-contained with bundled ffmpeg).
#
#   ./scripts/install-mac.sh              # rebuild, then install
#   ./scripts/install-mac.sh --no-build   # install the existing build/bin app
#
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export PATH="$PATH:$(go env GOPATH 2>/dev/null)/bin"

APP="$ROOT/build/bin/VideoForge.app"
DEST="/Applications/VideoForge.app"

if [ "${1:-}" != "--no-build" ]; then
  if ! command -v wails >/dev/null 2>&1; then
    echo "ERROR: wails CLI not found. Install it:" >&2
    echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
    exit 1
  fi
  echo "-> building (wails build -clean)..."
  wails build -clean
fi

if [ ! -d "$APP" ]; then
  echo "ERROR: $APP not found. Run without --no-build to build it first." >&2
  exit 1
fi

echo "-> quitting any running instance..."
pkill -f "VideoForge.app/Contents/MacOS/VideoForge" 2>/dev/null || true
sleep 1

echo "-> installing to $DEST ..."
rm -rf "$DEST"
cp -R "$APP" "$DEST"

echo "-> indexing for Spotlight..."
mdimport -i "$DEST" 2>/dev/null || true

echo "OK. Launch via Spotlight (Cmd-Space -> VideoForge) or: open -a VideoForge"
open -a "$DEST" 2>/dev/null || true

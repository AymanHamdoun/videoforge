#!/usr/bin/env bash
#
# Post-build hook: copies the bundled ffmpeg/ffprobe into the packaged macOS .app
# so the shipped application is fully self-contained. Invoked by wails.json's
# postBuildHooks (and safe to run by hand). No-op on non-macOS platforms, where
# the NSIS installer copies the binaries next to the .exe instead.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
platform="${1:-}"

case "$platform" in
  darwin*)
    app="$ROOT/build/bin/VideoForge.app"
    if [ ! -d "$app" ]; then
      echo "bundle-ffmpeg: no .app at $app — skipping"
      exit 0
    fi
    src="$ROOT/resources/bin/darwin"
    if [ ! -f "$src/ffmpeg" ] || [ ! -f "$src/ffprobe" ]; then
      echo "bundle-ffmpeg: missing $src/ffmpeg|ffprobe — run scripts/fetch-ffmpeg.sh" >&2
      exit 1
    fi
    dest="$app/Contents/Resources/bin"
    mkdir -p "$dest"
    cp "$src/ffmpeg" "$src/ffprobe" "$dest/"
    chmod +x "$dest/ffmpeg" "$dest/ffprobe"
    echo "✓ bundled ffmpeg + ffprobe into $dest"
    ;;
  *)
    echo "bundle-ffmpeg: nothing to do for platform '$platform'"
    ;;
esac

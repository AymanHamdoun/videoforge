#!/usr/bin/env bash
#
# Downloads static ffmpeg + ffprobe into resources/bin/<os>/ so they can be
# bundled into the packaged app. Run this once before `wails build` (and in CI).
#
#   ./scripts/fetch-ffmpeg.sh            # auto-detect host OS
#   ./scripts/fetch-ffmpeg.sh windows    # fetch Windows .exe binaries (e.g. for cross-build)
#
# macOS binaries come from evermeet.cx (x86_64 static; runs natively on Intel and
# under Rosetta 2 on Apple Silicon). For a native arm64 / universal build, swap the
# source for an arm64 static build (e.g. osxexperts.net) — the layout is identical.
# Linux/Windows hosts pull from the BtbN static-build releases.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

fetch_darwin() {
  local dest="$ROOT/resources/bin/darwin"
  mkdir -p "$dest"
  local tmp; tmp="$(mktemp -d)"
  echo "→ downloading macOS ffmpeg + ffprobe (evermeet.cx)…"
  curl -fsSL -o "$tmp/ffmpeg.zip"  "https://evermeet.cx/ffmpeg/getrelease/ffmpeg/zip"
  curl -fsSL -o "$tmp/ffprobe.zip" "https://evermeet.cx/ffmpeg/getrelease/ffprobe/zip"
  unzip -oq "$tmp/ffmpeg.zip"  -d "$dest"
  unzip -oq "$tmp/ffprobe.zip" -d "$dest"
  chmod +x "$dest/ffmpeg" "$dest/ffprobe"
  xattr -dr com.apple.quarantine "$dest/ffmpeg" "$dest/ffprobe" 2>/dev/null || true
  rm -rf "$tmp"
  echo "✓ macOS binaries in $dest"
  "$dest/ffmpeg" -version | head -1
}

fetch_linux() {
  local dest="$ROOT/resources/bin/linux"
  mkdir -p "$dest"
  local tmp; tmp="$(mktemp -d)"
  local url="https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz"
  echo "→ downloading Linux ffmpeg + ffprobe (BtbN)…"
  curl -fsSL -o "$tmp/ff.tar.xz" "$url"
  tar -xJf "$tmp/ff.tar.xz" -C "$tmp"
  local bindir; bindir="$(find "$tmp" -maxdepth 2 -type d -name bin | head -1)"
  cp "$bindir/ffmpeg" "$bindir/ffprobe" "$dest/"
  chmod +x "$dest/ffmpeg" "$dest/ffprobe"
  rm -rf "$tmp"
  echo "✓ Linux binaries in $dest"
}

fetch_windows() {
  local dest="$ROOT/resources/bin/windows"
  mkdir -p "$dest"
  local tmp; tmp="$(mktemp -d)"
  echo "→ downloading Windows ffmpeg + ffprobe (gyan.dev)…"
  curl -fsSL -o "$tmp/ff.zip" "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip"
  unzip -oq "$tmp/ff.zip" -d "$tmp"
  local bindir; bindir="$(find "$tmp" -maxdepth 2 -type d -name bin | head -1)"
  cp "$bindir/ffmpeg.exe" "$bindir/ffprobe.exe" "$dest/"
  rm -rf "$tmp"
  echo "✓ Windows binaries in $dest"
}

# Allow an explicit target OS (handy for cross-builds / CI); default to host.
target="${1:-}"
if [ -z "$target" ]; then
  case "$(uname -s)" in
    Darwin) target=darwin ;;
    Linux)  target=linux ;;
    *) echo "Unsupported host: $(uname -s). For Windows run scripts/fetch-ffmpeg.ps1" >&2; exit 1 ;;
  esac
fi

case "$target" in
  darwin)  fetch_darwin ;;
  linux)   fetch_linux ;;
  windows) fetch_windows ;;
  *) echo "Unknown target '$target' (use: darwin | linux | windows)" >&2; exit 1 ;;
esac

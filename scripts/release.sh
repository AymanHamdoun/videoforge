#!/usr/bin/env bash
#
# Cuts a release locally and publishes it to GitHub Releases.
#
#   ./scripts/release.sh                # uses the tag on HEAD (must be on a tagged commit)
#   ./scripts/release.sh v0.1.0         # explicit version
#
# Requires:
#   - macOS host (for the .dmg build; Windows is cross-compiled from here)
#   - wails on PATH (export PATH="$PATH:$(go env GOPATH)/bin")
#   - gh (GitHub CLI) installed and authenticated: `brew install gh && gh auth login`
#
# The .dmg / .exe are attached as GitHub Release assets, so the web app's
# /api/download/[platform] can redirect to them via the GitHub Releases API
# (set GITHUB_RELEASES_REPO in the web app's env).

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-$(git describe --tags --exact-match 2>/dev/null || true)}"
if [ -z "$VERSION" ]; then
  echo "usage: scripts/release.sh vX.Y.Z   (or run on a tagged commit)" >&2
  exit 1
fi
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  echo "error: version '$VERSION' must look like v0.1.0" >&2
  exit 1
fi

command -v wails >/dev/null || { echo "error: wails not on PATH (try: export PATH=\"\$PATH:\$(go env GOPATH)/bin\")" >&2; exit 1; }
command -v hdiutil >/dev/null || { echo "error: hdiutil not found (this script needs macOS)" >&2; exit 1; }
command -v gh >/dev/null || { echo "error: gh (GitHub CLI) not found — brew install gh && gh auth login" >&2; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "error: gh is not authenticated — run: gh auth login" >&2; exit 1; }

DIST="$ROOT/dist"
rm -rf "$DIST"
mkdir -p "$DIST"

MAC_DMG="VideoForge-macos-universal.dmg"
WIN_EXE="VideoForge-windows-amd64-installer.exe"

echo "→ building macOS universal .app"
./scripts/fetch-ffmpeg.sh darwin
wails build -platform darwin/universal -clean
hdiutil create -volname VideoForge \
  -srcfolder "build/bin/VideoForge.app" \
  -ov -format UDZO "$DIST/$MAC_DMG"

echo "→ building Windows NSIS installer"
./scripts/fetch-ffmpeg.sh windows
wails build -platform windows/amd64 -nsis -clean
mv "build/bin/VideoForge-amd64-installer.exe" "$DIST/$WIN_EXE"

# Publish to GitHub Releases (gh infers the repo from this checkout's origin).
# Idempotent: create the release if it's new, otherwise just (re)upload assets.
if gh release view "$VERSION" >/dev/null 2>&1; then
  echo "→ release $VERSION exists — uploading assets"
  gh release upload "$VERSION" "$DIST/$MAC_DMG" "$DIST/$WIN_EXE" --clobber
else
  echo "→ creating release $VERSION"
  gh release create "$VERSION" "$DIST/$MAC_DMG" "$DIST/$WIN_EXE" \
    --title "$VERSION" \
    --notes "VideoForge $VERSION"
fi

echo
echo "✓ released $VERSION"
gh release view "$VERSION" --json url --jq '"  " + .url' 2>/dev/null || true

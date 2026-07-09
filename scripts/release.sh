#!/usr/bin/env bash
#
# Cuts a release locally and publishes it to a GitLab project's Releases.
#
#   ./scripts/release.sh                # uses the tag on HEAD (must be on a tagged commit)
#   ./scripts/release.sh v0.1.0         # explicit version
#
# Requires:
#   - macOS host (for the .dmg build; Windows is cross-compiled from here)
#   - wails on PATH (export PATH="$PATH:$(go env GOPATH)/bin")
#   - GITLAB_TOKEN env var: Personal Access Token with `api` scope
#   - GITLAB_PROJECT_ID env var: numeric project id from the gitlab project page
#
# Optional:
#   - GITLAB_HOST (defaults to https://gitlab.com)
#
# Artifacts are uploaded to the project's Generic Package Registry and linked
# from the Release, so the web app's /api/download/[platform] can resolve them
# via the Releases API.

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

: "${GITLAB_TOKEN:?need GITLAB_TOKEN (Personal Access Token with 'api' scope)}"
: "${GITLAB_PROJECT_ID:?need GITLAB_PROJECT_ID (numeric project id)}"
GITLAB_HOST="${GITLAB_HOST:-https://gitlab.com}"

command -v wails >/dev/null || { echo "error: wails not on PATH (try: export PATH=\"\$PATH:\$(go env GOPATH)/bin\")" >&2; exit 1; }
command -v hdiutil >/dev/null || { echo "error: hdiutil not found (this script needs macOS)" >&2; exit 1; }

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

API="$GITLAB_HOST/api/v4"
PKG="videoforge"
PKG_VER="${VERSION#v}"

upload() {
  local path="$1"; local name; name="$(basename "$path")"
  local url="$API/projects/$GITLAB_PROJECT_ID/packages/generic/$PKG/$PKG_VER/$name"
  echo "→ uploading $name" >&2
  curl --fail-with-body -sS \
    --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
    --upload-file "$path" \
    "$url" >/dev/null
  printf '%s' "$url"
}

MAC_URL="$(upload "$DIST/$MAC_DMG")"
WIN_URL="$(upload "$DIST/$WIN_EXE")"

echo "→ creating release $VERSION"
release_body() {
  cat <<EOF
{
  "name": "$VERSION",
  "tag_name": "$VERSION",
  "description": "VideoForge $VERSION",
  "assets": {
    "links": [
      { "name": "$MAC_DMG", "url": "$MAC_URL", "link_type": "package" },
      { "name": "$WIN_EXE", "url": "$WIN_URL", "link_type": "package" }
    ]
  }
}
EOF
}

curl --fail-with-body -sS -X POST \
  --header "PRIVATE-TOKEN: $GITLAB_TOKEN" \
  --header "Content-Type: application/json" \
  --data "$(release_body)" \
  "$API/projects/$GITLAB_PROJECT_ID/releases" >/dev/null

echo
echo "✓ released $VERSION"
echo "  mac: $MAC_URL"
echo "  win: $WIN_URL"

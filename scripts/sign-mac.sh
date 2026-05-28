#!/usr/bin/env bash
#
# Signs, notarizes, and packages the macOS app into a distributable .dmg that
# opens with no Gatekeeper warning on other Macs.
#
# One-time prerequisites (need your Apple login — see README / chat):
#   1. A "Developer ID Application" certificate in your login keychain
#      (Xcode → Settings → Accounts → Manage Certificates → + → Developer ID Application).
#   2. A stored notarytool credential profile named $NOTARY_PROFILE:
#        xcrun notarytool store-credentials VF_NOTARY \
#          --apple-id you@example.com --team-id YOURTEAMID --password <app-specific-password>
#
# Then just: ./scripts/sign-mac.sh   (after `wails build -clean`)
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
APP="$ROOT/build/bin/VideoForge.app"
DMG="$ROOT/build/bin/VideoForge.dmg"
NOTARY_PROFILE="${NOTARY_PROFILE:-VF_NOTARY}"

# Auto-detect the Developer ID Application identity unless one is provided.
SIGN_IDENTITY="${SIGN_IDENTITY:-$(security find-identity -v -p codesigning \
  | awk -F'"' '/Developer ID Application/{print $2; exit}')}"
[ -n "$SIGN_IDENTITY" ] || { echo "ERROR: no 'Developer ID Application' identity in keychain"; exit 1; }
[ -d "$APP" ] || { echo "ERROR: $APP not found — run 'wails build -clean' first"; exit 1; }
echo "Signing identity: $SIGN_IDENTITY"

# 1) Sign the bundled third-party binaries first (hardened runtime + timestamp).
for bin in ffmpeg ffprobe; do
  echo "→ signing $bin"
  codesign --force --timestamp --options runtime \
    --sign "$SIGN_IDENTITY" "$APP/Contents/Resources/bin/$bin"
done

# 2) Sign the app bundle (deep covers the Wails main binary).
echo "→ signing app bundle"
codesign --force --timestamp --options runtime --deep \
  --sign "$SIGN_IDENTITY" "$APP"
codesign --verify --strict --verbose=2 "$APP"

# 3) Build a compressed DMG.
echo "→ creating dmg"
rm -f "$DMG"
hdiutil create -volname VideoForge -srcfolder "$APP" -ov -format UDZO "$DMG"
codesign --force --timestamp --sign "$SIGN_IDENTITY" "$DMG"

# 4) Notarize and staple.
echo "→ notarizing (a few minutes)…"
xcrun notarytool submit "$DMG" --keychain-profile "$NOTARY_PROFILE" --wait
xcrun stapler staple "$APP"
xcrun stapler staple "$DMG"

# 5) Verify Gatekeeper will accept it.
echo "→ verifying"
spctl -a -t exec -vv "$APP" || true
echo "✅ Done: $DMG"

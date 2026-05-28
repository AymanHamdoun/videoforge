#!/usr/bin/env bash
#
# Interactive license issuer. Prompts for name, email, and expiry, then prints a
# signed license key (using the private key in .secrets/license_ed25519.key).
#
#   ./scripts/make-license.sh
#
# Override the key location with: LICENSE_KEY_FILE=/path/to/key ./scripts/make-license.sh
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

KEY_FILE="${LICENSE_KEY_FILE:-$ROOT/.secrets/license_ed25519.key}"
if [ ! -f "$KEY_FILE" ]; then
  echo "✗ Private key not found at: $KEY_FILE" >&2
  echo "  Generate a keypair first:  go run ./cmd/license-gen keygen" >&2
  exit 1
fi

read -r -p "Licensee name:            " NAME
read -r -p "Licensee email:           " EMAIL
read -r -p "Expiry in days (0=never): " DAYS
DAYS="${DAYS:-0}"

if ! [[ "$DAYS" =~ ^[0-9]+$ ]]; then
  echo "✗ Expiry must be a non-negative whole number (0 for never)." >&2
  exit 1
fi

ARGS=(--name "$NAME" --email "$EMAIL")
if [ "$DAYS" -gt 0 ]; then
  ARGS+=(--days "$DAYS")
fi

echo
echo "──────────────────────────────────────────────"
if [ "$DAYS" -gt 0 ]; then
  echo "License for $NAME <$EMAIL> — expires in $DAYS day(s):"
else
  echo "License for $NAME <$EMAIL> — perpetual:"
fi
echo "──────────────────────────────────────────────"
LICENSE_PRIVATE_KEY="$(cat "$KEY_FILE")" go run ./cmd/license-gen issue "${ARGS[@]}"
echo "──────────────────────────────────────────────"

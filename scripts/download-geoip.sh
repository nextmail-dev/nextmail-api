#!/usr/bin/env bash
# Download and install the MaxMind GeoLite2-Country database.
#
# Requires a free MaxMind account (https://www.maxmind.com/en/geolite2/signup):
#   GEOIP_ACCOUNT_ID   - your MaxMind account ID
#   GEOIP_LICENSE_KEY  - your MaxMind license key
#
# Optional:
#   DATA_DIR - output directory (default: ./data)
set -euo pipefail

ACCOUNT_ID="${GEOIP_ACCOUNT_ID:?GEOIP_ACCOUNT_ID env var is required}"
LICENSE_KEY="${GEOIP_LICENSE_KEY:?GEOIP_LICENSE_KEY env var is required}"
DATA_DIR="${DATA_DIR:-./data}"

mkdir -p "$DATA_DIR"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading GeoLite2-Country..."
curl -fsSL -u "$ACCOUNT_ID:$LICENSE_KEY" \
  "https://download.maxmind.com/geoip/databases/GeoLite2-Country/download?suffix=tar.gz" \
  -o "$TMP_DIR/GeoLite2-Country.tar.gz"

echo "Extracting..."
tar -xzf "$TMP_DIR/GeoLite2-Country.tar.gz" -C "$TMP_DIR"

MMDB="$(find "$TMP_DIR" -name '*.mmdb' | head -n1)"
if [ -z "$MMDB" ]; then
  echo "error: no .mmdb file found in archive" >&2
  exit 1
fi

mv "$MMDB" "$DATA_DIR/GeoLite2-Country.mmdb"
echo "Installed: $DATA_DIR/GeoLite2-Country.mmdb"

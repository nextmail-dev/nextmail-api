#!/usr/bin/env bash
# Cross-compile and package nextmail-api for multiple platforms.
#
# Produces, under dist/:
#   nextmail-api_<version>_<os>_<arch>.tar.gz   (linux / darwin)
#   nextmail-api_<version>_<os>_<arch>.zip      (windows)
#   checksums.txt                               SHA-256 of every archive
#
# Environment variables:
#   VERSION   build version label (default: git describe, or "dev")
#   DIST_DIR  output directory (default: dist)
set -euo pipefail

MODULE="nextmail-api"
CMD="./cmd/server"
DIST_DIR="${DIST_DIR:-dist}"

# Target platforms, expressed as GOOS/GOARCH. Edit this list to suit.
TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

# ldflags: strip symbols + DWARF (-s -w) for smaller binaries, and stamp the
# version into internal/version.Version (a no-op until that package exists).
if [ -z "${VERSION:-}" ]; then
  VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
fi
LDFLAGS="-s -w -X ${MODULE}/internal/version.Version=${VERSION}"

# Pick a SHA-256 tool: sha256sum on Linux, shasum on macOS.
if command -v sha256sum >/dev/null 2>&1; then
  SHA256=(sha256sum)
else
  SHA256=(shasum -a 256)
fi

# --- preflight --------------------------------------------------------------
command -v go >/dev/null 2>&1 || { echo "error: go toolchain not found in PATH" >&2; exit 1; }

need_zip=0
for t in "${TARGETS[@]}"; do
  case "$t" in windows/*) need_zip=1 ;; esac
done
if [ "$need_zip" = "1" ]; then
  command -v zip >/dev/null 2>&1 || { echo "error: 'zip' is required for windows targets (apt install zip)" >&2; exit 1; }
fi

# --- build ------------------------------------------------------------------
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"
DIST_DIR="$(cd "$DIST_DIR" && pwd)"

checksums="$DIST_DIR/checksums.txt"
: > "$checksums"

for target in "${TARGETS[@]}"; do
  os="${target%%/*}"
  arch="${target##*/}"

  ext="tar.gz"
  [ "$os" = "windows" ] && ext="zip"

  bin="$MODULE"
  [ "$os" = "windows" ] && bin="${MODULE}.exe"

  archive_name="${MODULE}_${VERSION}_${os}_${arch}"
  archive="${DIST_DIR}/${archive_name}.${ext}"
  staging="$(mktemp -d)"

  echo "==> Building ${os}/${arch}"

  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath -ldflags "$LDFLAGS" -o "${staging}/${bin}" "$CMD"

  [ -f README.md ] && cp README.md "$staging/"

  if [ "$os" = "windows" ]; then
    ( cd "$staging" && zip -rq "$archive" . )
  else
    tar -C "$staging" -czf "$archive" .
  fi

  rm -rf "$staging"

  ( cd "$DIST_DIR" && "${SHA256[@]}" "${archive_name}.${ext}" ) >> "$checksums"
  echo "    -> ${archive_name}.${ext}"
done

echo
echo "Built ${#TARGETS[@]} targets. Artifacts in ${DIST_DIR}/:"
ls -1 "$DIST_DIR"

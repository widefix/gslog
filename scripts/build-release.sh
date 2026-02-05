#!/usr/bin/env bash
# Build git-squash-tree for all platforms and package for GitHub Releases.
# Usage: ./scripts/build-release.sh [version]
#   version defaults to "dev" if not set (e.g. use v0.1.0 for releases).
#
# macOS signing and notarization (optional, for Gatekeeper):
#   Set SIGNING_IDENTITY and NOTARY_KEYCHAIN_PROFILE before running.
#   See docs/apple-signing.md for setup.
#   Without these, macOS binaries build but are unsigned (users may need xattr -d).

set -e

VERSION="${1:-dev}"
DIST="dist"
BINARY_NAME="git-squash-tree"
BUILD_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$BUILD_DIR"
mkdir -p "$DIST"

# Sign and notarize a macOS binary. Requires SIGNING_IDENTITY and NOTARY_KEYCHAIN_PROFILE.
sign_and_notarize() {
  local binary="$1"
  if [ -z "${SIGNING_IDENTITY:-}" ] || [ -z "${NOTARY_KEYCHAIN_PROFILE:-}" ]; then
    echo "  Skipping signing (set SIGNING_IDENTITY and NOTARY_KEYCHAIN_PROFILE to enable)"
    return
  fi

  echo "  Signing..."
  codesign --force --options runtime --timestamp --sign "$SIGNING_IDENTITY" "$binary"

  echo "  Submitting for notarization..."
  local zip_path="${binary}.zip"
  zip -j "$zip_path" "$binary"
  xcrun notarytool submit "$zip_path" --keychain-profile "$NOTARY_KEYCHAIN_PROFILE" --wait
  rm -f "$zip_path"

  # Stapling only works on .app, .dmg, .pkg — not raw binaries. Notarization is
  # already recorded for this binary; Gatekeeper verifies it online via the signature.
  echo "  Notarized (online verification; stapling not supported for CLI binaries)."
}

# Build a single target: GOOS GOARCH suffix (e.g. Darwin_arm64) [.exe for windows]
build() {
  local goos="$1"
  local goarch="$2"
  local suffix="$3"
  local ext="${4:-}"

  local output="$DIST/${BINARY_NAME}${ext}"
  echo "Building $goos/$goarch -> $suffix"
  GOOS="$goos" GOARCH="$goarch" go build -ldflags "-s -w" -o "$output" ./cmd/git-squash-tree

  if [ "$goos" = "darwin" ]; then
    sign_and_notarize "$output"
  fi

  if [ "$goos" = "windows" ]; then
    zip -j "$DIST/${BINARY_NAME}_${suffix}.zip" "$output"
  else
    tar -czf "$DIST/${BINARY_NAME}_${suffix}.tar.gz" -C "$DIST" "${BINARY_NAME}${ext}"
  fi
  rm -f "$output"
}

build darwin amd64  "Darwin_x86_64"
build darwin arm64  "Darwin_arm64"
build linux amd64   "Linux_x86_64"
build linux arm64   "Linux_arm64"
build windows amd64 "Windows_x86_64" ".exe"

echo ""
echo "Done. Artifacts in $DIST/:"
ls -la "$DIST"

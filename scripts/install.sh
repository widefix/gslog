#!/usr/bin/env bash

set -e

REPO="${REPO:-widefix/squash-tree}"
VERSION="${1:-${VERSION:-latest}}"

# Detect OS and arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) OS="Darwin" ;;
  Linux)  OS="Linux" ;;
  *)
    echo "Unsupported OS: $OS (this script supports macOS and Linux only)"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Resolve version
if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -sSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' | head -1)"
  [ -z "$VERSION" ] && { echo "Could not fetch latest version. Check that $REPO has releases."; exit 1; }
fi

ASSET="git-squash-tree_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Installing git-squash-tree $VERSION ($OS/$ARCH)..."

# Download and extract
tmp="$(mktemp -d)"
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT

if ! curl -sSLf "$URL" -o "$tmp/archive.tar.gz"; then
  echo "Download failed. Check that $URL exists."
  exit 1
fi

tar -xzf "$tmp/archive.tar.gz" -C "$tmp"

# Install location
if [ -w /usr/local/bin ] 2>/dev/null; then
  INSTALL_DIR="/usr/local/bin"
else
  INSTALL_DIR="$HOME/bin"
  mkdir -p "$INSTALL_DIR"
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    PATH_LINE='export PATH="$HOME/bin:$PATH"'
    for rc in .zshrc .bashrc; do
      rcfile="$HOME/$rc"
      if [ -f "$rcfile" ]; then
        if ! grep -qF "$HOME/bin" "$rcfile" 2>/dev/null; then
          echo "" >> "$rcfile"
          echo "# squash-tree" >> "$rcfile"
          echo "$PATH_LINE" >> "$rcfile"
          echo "Added $INSTALL_DIR to PATH in $rc"
        fi
      else
        echo "$PATH_LINE" >> "$rcfile"
        echo "Created $rc with PATH entry"
      fi
    done
    echo "Run 'source ~/.zshrc' or 'source ~/.bashrc' (or open a new terminal) to use git-squash-tree."
  fi
fi

mv "$tmp/git-squash-tree" "$INSTALL_DIR/git-squash-tree"
chmod +x "$INSTALL_DIR/git-squash-tree"

# Git alias
git config --global alias.squash-tree '! git-squash-tree'

echo ""
echo "Installed to $INSTALL_DIR/git-squash-tree"
echo ""
echo "Next step: run in a repository:"
echo "  git squash-tree init           # this repo only"
echo "  git squash-tree init --global  # all repos"

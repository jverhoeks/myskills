#!/bin/bash
set -euo pipefail

REPO="jverhoeks/myskills"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

echo "Fetching latest release..."
LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release. Is the repo public or do you need GITHUB_TOKEN?" >&2
  exit 1
fi

FILENAME="myskills_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"
echo "Downloading myskills ${LATEST} (${OS}/${ARCH})..."

TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

curl -sL "$URL" -o "${TMP}/${FILENAME}"
tar -xzf "${TMP}/${FILENAME}" -C "$TMP"

echo "Installing to ${INSTALL_DIR}/myskills..."
install -m 755 "${TMP}/myskills" "${INSTALL_DIR}/myskills"
echo "✓ myskills ${LATEST} installed"

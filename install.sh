#!/usr/bin/env bash
set -euo pipefail

REPO="miraccan00/go-precommit"
BINARY="go-precommit"

# ── OS detection ──────────────────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  darwin)             OS="darwin" ;;
  linux)              OS="linux" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "error: unsupported OS: $OS" >&2
    exit 1
    ;;
esac

# ── Architecture detection ────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *)
    echo "error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# ── Resolve latest release version ───────────────────────────────────────────
VERSION="${GOPRECOMMIT_VERSION:-}"
if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
fi

if [ -z "$VERSION" ]; then
  echo "error: could not determine latest version (check your internet connection or set GOPRECOMMIT_VERSION)" >&2
  exit 1
fi

# ── Build download URL ────────────────────────────────────────────────────────
FILENAME="${BINARY}_${OS}_${ARCH}"
[ "$OS" = "windows" ] && FILENAME="${FILENAME}.exe"

URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Installing ${BINARY} ${VERSION} (${OS}/${ARCH})..."

# ── Download ──────────────────────────────────────────────────────────────────
TMP="$(mktemp)"
trap 'rm -f "$TMP"' EXIT

if ! curl -fsSL "$URL" -o "$TMP"; then
  echo "error: download failed: $URL" >&2
  echo "       Check that release ${VERSION} exists and includes a ${FILENAME} asset." >&2
  exit 1
fi
chmod +x "$TMP"

# ── Install ───────────────────────────────────────────────────────────────────
# Prefer ~/.local/bin (no sudo) when /usr/local/bin is not writable.
if [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
elif [ -w "${HOME}/.local/bin" ] || mkdir -p "${HOME}/.local/bin" 2>/dev/null; then
  INSTALL_DIR="${HOME}/.local/bin"
  mv "$TMP" "${INSTALL_DIR}/${BINARY}"
else
  INSTALL_DIR="/usr/local/bin"
  echo "Installing to ${INSTALL_DIR} requires sudo..."
  sudo mv "$TMP" "${INSTALL_DIR}/${BINARY}"
fi

echo ""
echo "${BINARY} ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo ""
echo "Next step — activate globally for all git repositories:"
echo ""
echo "  ${BINARY} global-install"
echo ""

# Warn if install dir is not on PATH
case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo "warning: ${INSTALL_DIR} is not in your PATH."
    echo "         Add the following to your shell profile (~/.zshrc, ~/.bashrc, etc.):"
    echo ""
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
    ;;
esac

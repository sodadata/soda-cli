#!/bin/sh
#
# Soda CLI installer
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/sodadata/soda-cli/main/install.sh | sh
#
# Options (environment variables):
#   SODACLI_VERSION   - specific version to install (default: latest)
#   SODACLI_INSTALL   - install directory (default: /usr/local/bin)
#

set -e

REPO="sodadata/soda-cli"
BINARY="sodacli"
INSTALL_DIR="${SODACLI_INSTALL:-/usr/local/bin}"

# ── Detect OS and architecture ────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "Unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

# ── Resolve version ──────────────────────────────────────────────────────────

if [ -z "$SODACLI_VERSION" ]; then
  SODACLI_VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')
  if [ -z "$SODACLI_VERSION" ]; then
    echo "Error: could not determine latest version. Set SODACLI_VERSION manually." >&2
    exit 1
  fi
fi

# Strip leading 'v' for the archive name
VERSION_NUM="${SODACLI_VERSION#v}"

echo "Installing ${BINARY} ${SODACLI_VERSION} (${OS}/${ARCH})..."

# ── Download and install ──────────────────────────────────────────────────────

EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${SODACLI_VERSION}/${ARCHIVE}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${URL}..."
curl -sSL -o "${TMP_DIR}/${ARCHIVE}" "$URL"

if [ ! -s "${TMP_DIR}/${ARCHIVE}" ]; then
  echo "Error: download failed or file is empty." >&2
  exit 1
fi

# Extract
cd "$TMP_DIR"
if [ "$EXT" = "tar.gz" ]; then
  tar xzf "$ARCHIVE"
elif [ "$EXT" = "zip" ]; then
  unzip -q "$ARCHIVE"
fi

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "Installed ${BINARY} ${SODACLI_VERSION} to ${INSTALL_DIR}/${BINARY}"
echo ""
"${INSTALL_DIR}/${BINARY}" version

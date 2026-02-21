#!/usr/bin/env sh
set -eu

REPO_OWNER="${REPO_OWNER:-naodEthiop}"
REPO_NAME="${REPO_NAME:-lalibela-cli}"
BINARY_NAME="${BINARY_NAME:-lalibela}"
VERSION="${VERSION:-latest}"

detect_os() {
  uname_out="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$uname_out" in
    linux*) echo "linux" ;;
    darwin*) echo "darwin" ;;
    msys*|mingw*|cygwin*) echo "windows" ;;
    *) echo "unsupported" ;;
  esac
}

detect_arch() {
  uname_arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
  case "$uname_arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
  echo "Unsupported OS/ARCH: $OS/$ARCH" >&2
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
fi

if [ -z "$VERSION" ]; then
  echo "Could not determine release version." >&2
  exit 1
fi

EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

ARCHIVE="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}.${EXT}"
BASE_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

curl -fL "${BASE_URL}/${ARCHIVE}" -o "${TMP_DIR}/${ARCHIVE}"
curl -fL "${BASE_URL}/checksums.txt" -o "${TMP_DIR}/checksums.txt"

EXPECTED="$(grep " ${ARCHIVE}\$" "${TMP_DIR}/checksums.txt" | awk '{print $1}')"
if [ -z "$EXPECTED" ]; then
  echo "Checksum not found for ${ARCHIVE}" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "${TMP_DIR}/${ARCHIVE}" | awk '{print $1}')"
else
  echo "No SHA256 command found (sha256sum or shasum required)." >&2
  exit 1
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Checksum verification failed for ${ARCHIVE}" >&2
  exit 1
fi

if [ "$EXT" = "zip" ]; then
  unzip -q "${TMP_DIR}/${ARCHIVE}" -d "$TMP_DIR"
else
  tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"
fi

BIN_PATH="${TMP_DIR}/${BINARY_NAME}"
if [ "$OS" = "windows" ]; then
  BIN_PATH="${TMP_DIR}/${BINARY_NAME}.exe"
fi

if [ ! -f "$BIN_PATH" ]; then
  BIN_PATH="$(find "$TMP_DIR" -type f \( -name "${BINARY_NAME}" -o -name "${BINARY_NAME}.exe" \) | head -n 1)"
fi
if [ -z "$BIN_PATH" ] || [ ! -f "$BIN_PATH" ]; then
  echo "Binary not found in release archive." >&2
  exit 1
fi

if [ "$OS" = "windows" ]; then
  INSTALL_DIR="${INSTALL_DIR:-$HOME/bin}"
  mkdir -p "$INSTALL_DIR"
  cp "$BIN_PATH" "${INSTALL_DIR}/${BINARY_NAME}.exe"
  echo "Installed ${BINARY_NAME}.exe to ${INSTALL_DIR}"
  echo "Add ${INSTALL_DIR} to PATH if needed."
  exit 0
fi

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TARGET="${INSTALL_DIR}/${BINARY_NAME}"

if [ "$(id -u)" -ne 0 ] && [ "$INSTALL_DIR" = "/usr/local/bin" ]; then
  sudo install -m 0755 "$BIN_PATH" "$TARGET"
else
  install -m 0755 "$BIN_PATH" "$TARGET"
fi

echo "Installed ${BINARY_NAME} to ${TARGET}"


#!/bin/bash
# NeoCode installer — downloads and installs the latest release.
#
# Usage:
#   curl -fsSL https://get.neocode.dev | bash
#   curl -fsSL https://get.neocode.dev | bash -s -- --version 2.0.0

set -euo pipefail

REPO="user/neocode"
BINARY="neocode"
INSTALL_DIR="${NEOCODE_INSTALL_DIR:-$HOME/.neocode/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[info]${NC} $1"; }
warn()  { echo -e "${YELLOW}[warn]${NC} $1"; }
error() { echo -e "${RED}[error]${NC} $1"; exit 1; }

# Detect OS
detect_os() {
  case "$(uname -s)" in
    Darwin*) echo "darwin" ;;
    Linux*)  echo "linux" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) error "Unsupported OS: $(uname -s)" ;;
  esac
}

# Detect architecture
detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) error "Unsupported architecture: $(uname -m)" ;;
  esac
}

# Get latest release version
get_latest_version() {
  local version
  version=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed -E 's/.*"v([^"]+)".*/\1/')
  if [ -z "$version" ]; then
    error "Failed to fetch latest version"
  fi
  echo "$version"
}

main() {
  local os arch version ext url tmp_file

  os=$(detect_os)
  arch=$(detect_arch)

  # Parse arguments
  version=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --version) version="$2"; shift 2 ;;
      --dir) INSTALL_DIR="$2"; shift 2 ;;
      *) shift ;;
    esac
  done

  if [ -z "$version" ]; then
    version=$(get_latest_version)
  fi

  info "Installing NeoCode ${version} for ${os}/${arch}..."

  ext=""
  if [ "$os" = "windows" ]; then
    ext=".exe"
  fi

  url="https://github.com/${REPO}/releases/download/v${version}/neocode-${os}-${arch}${ext}"
  tmp_file=$(mktemp)

  info "Downloading ${url}..."
  if ! curl -sL -o "$tmp_file" "$url"; then
    rm -f "$tmp_file"
    error "Download failed. Check if version ${version} exists."
  fi

  mkdir -p "$INSTALL_DIR"

  local install_path="${INSTALL_DIR}/${BINARY}${ext}"
  mv "$tmp_file" "$install_path"
  chmod +x "$install_path"

  info "Installed to ${install_path}"

  # Add to PATH if not already there
  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    local shell_rc=""
    case "$SHELL" in
      */zsh)  shell_rc="$HOME/.zshrc" ;;
      */bash) shell_rc="$HOME/.bashrc" ;;
      *)      shell_rc="$HOME/.profile" ;;
    esac

    echo "" >> "$shell_rc"
    echo "# NeoCode" >> "$shell_rc"
    echo "export PATH=\"${INSTALL_DIR}:\$PATH\"" >> "$shell_rc"

    info "Added ${INSTALL_DIR} to PATH in ${shell_rc}"
    warn "Run 'source ${shell_rc}' or restart your terminal to use neocode"
  fi

  info "✓ NeoCode ${version} installed successfully!"
  echo ""
  echo "  Run 'neocode' to start, or 'neocode --help' for usage."
  echo "  Configure: export NEOCODE_API_KEY=your-key"
  echo ""
}

main "$@"

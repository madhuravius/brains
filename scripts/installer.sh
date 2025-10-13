#!/usr/bin/env bash
set -Eeuo pipefail

OWNER="madhuravius"
REPO="brains"
BINARY="brains"

# Defaults
VERSION="${VERSION:-}"

usage() {
  cat <<EOF
brains installer

Usage:
  install.sh [--version vX.Y.Z]

Options:
  --version, -v  Version tag to install (e.g., v1.2.3). Default: latest release
  -h, --help     Show this help

Installation location:
  - /usr/local/bin (with sudo if needed)
  - ~/.local/bin (fallback if sudo unavailable)
EOF
}

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    -v|--version)
      VERSION="${2:-}"
      if [[ -z "$VERSION" ]]; then
        echo "Error: --version requires a value" >&2
        usage
        exit 1
      fi
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

have() { command -v "$1" >/dev/null 2>&1; }

# Cleanup temp directory on exit
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# Detect OS and architecture
detect_platform() {
  local os arch
  
  case "$(uname -s)" in
    Linux)   os="linux" ;;
    Darwin)  os="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) os="windows" ;;
    *) echo "Unsupported OS: $(uname -s)"; exit 1 ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac

  echo "$os $arch"
}

read -r OS ARCH < <(detect_platform)
EXT="tar.gz"
[[ "$OS" == "windows" ]] && EXT="zip"

# Resolve latest version if not specified
resolve_version() {
  if [[ -n "$VERSION" ]]; then
    echo "$VERSION"
    return
  fi

  echo "Resolving latest release tag..." >&2
  
  # Try GitHub API first
  local tag
  tag="$(curl -fsSL "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" 2>/dev/null \
    | grep -o '"tag_name":[[:space:]]*"[^"]*"' \
    | head -n1 \
    | cut -d'"' -f4 || true)"
  
  if [[ -z "$tag" ]]; then
    # Fallback to redirect
    tag="$(curl -fsIL "https://github.com/${OWNER}/${REPO}/releases/latest" \
      | grep -i '^location:' \
      | tail -n1 \
      | sed 's/.*\/tag\/\([^[:space:]]*\).*/\1/' \
      | tr -d '\r' || true)"
  fi

  if [[ -z "$tag" ]]; then
    echo "Could not determine latest release. Specify --version vX.Y.Z" >&2
    exit 1
  fi

  echo "$tag"
}

VERSION="$(resolve_version)"

# Find release asset
find_asset() {
  local base_url="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
  local os_title="$(tr '[:lower:]' '[:upper:]' <<< "${OS:0:1}")${OS:1}"
  
  # Try different naming conventions
  local candidates=(
    "${REPO}_${VERSION}_${os_title}_${ARCH}.${EXT}"
    "${REPO}_${VERSION}_${OS}_${ARCH}.${EXT}"
    "${REPO}_${VERSION}_${os_title}_x86_64.${EXT}"
    "${REPO}_${VERSION}_${OS}_x86_64.${EXT}"
  )

  # Try API first
  local asset
  asset="$(curl -fsSL "https://api.github.com/repos/${OWNER}/${REPO}/releases/tags/${VERSION}" \
    | grep -o '"name":[[:space:]]*"[^"]*"' \
    | cut -d'"' -f4 \
    | grep -E "_${OS}_${ARCH}\.(tar\.gz|zip)$" \
    | head -n1 || true)"

  if [[ -n "$asset" ]]; then
    echo "$asset"
    return
  fi

  # Fallback: probe candidates
  for name in "${candidates[@]}"; do
    if curl -fsSI "${base_url}/${name}" >/dev/null 2>&1; then
      echo "$name"
      return
    fi
  done

  echo "Could not find release asset for ${OS}/${ARCH}" >&2
  printf '  Tried: %s\n' "${candidates[@]}" >&2
  exit 1
}

ASSET="$(find_asset)"
echo "Installing ${REPO} ${VERSION} for ${OS}/${ARCH}..."

# Download asset
archive="${tmpdir}/${ASSET}"
echo "Downloading ${ASSET}..."
curl -fsSL -o "$archive" "https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${ASSET}"

# Extract archive
workdir="${tmpdir}/extract"
mkdir -p "$workdir"

case "$EXT" in
  tar.gz) tar -xzf "$archive" -C "$workdir" ;;
  zip)    unzip -q "$archive" -d "$workdir" ;;
esac

# Find binary
bin_path="$(find "$workdir" -type f -name "$BINARY" -print -quit)"
if [[ -z "$bin_path" ]]; then
  echo "Could not locate '${BINARY}' in archive" >&2
  exit 1
fi

chmod +x "$bin_path"

# Install to PATH
install_binary() {
  local src="$1"
  
  # Try /usr/local/bin first
  if [[ -w /usr/local/bin ]]; then
    cp -f "$src" "/usr/local/bin/${BINARY}"
    echo "Installed to /usr/local/bin/${BINARY}"
    return 0
  fi
  
  # Try with sudo
  if have sudo && sudo -n true 2>/dev/null; then
    sudo cp -f "$src" "/usr/local/bin/${BINARY}"
    echo "Installed to /usr/local/bin/${BINARY}"
    return 0
  fi
  
  if have sudo; then
    echo "Installing to /usr/local/bin requires sudo..."
    if sudo cp -f "$src" "/usr/local/bin/${BINARY}"; then
      echo "Installed to /usr/local/bin/${BINARY}"
      return 0
    fi
  fi
  
  # Fallback to ~/.local/bin
  local user_bin="${HOME}/.local/bin"
  mkdir -p "$user_bin"
  cp -f "$src" "$user_bin/${BINARY}"
  echo "Installed to ${user_bin}/${BINARY}"
  
  # Check if on PATH
  if [[ ":$PATH:" != *":${user_bin}:"* ]]; then
    echo ""
    echo "⚠️  ${user_bin} is not on your PATH"
    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "  export PATH=\"${user_bin}:\$PATH\""
  fi
}

install_binary "$bin_path"
echo ""
echo "✓ Installation complete. Try: ${BINARY} --help"

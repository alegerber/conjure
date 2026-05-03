#!/usr/bin/env bash
#
# conjure installer
#
# Downloads the latest (or pinned) release archive for the current
# OS/architecture from GitHub Releases, extracts the binary, and places
# it under $BIN_DIR (default ~/.local/bin) so it stays on $PATH.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/alegerber/conjure/main/install.sh | bash
#   ./install.sh                           # install latest release
#   VERSION=v1.0.0 ./install.sh            # pin a specific version
#   BIN_DIR=~/bin ./install.sh             # custom install directory
#   ./install.sh --uninstall               # remove the installed binary
#

set -euo pipefail

REPO="alegerber/conjure"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
CLI_TARGET="$BIN_DIR/conjure"
VERSION="${VERSION:-}"

# Stale oh-my-zsh plugin from the 0.1.x bash era — cleaned up if present.
OMZ_PLUGIN_DIR="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/conjure"
OMZ_PLUGIN_FILE="$OMZ_PLUGIN_DIR/conjure.plugin.zsh"

green()  { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
err()    { printf '\033[31m%s\033[0m\n' "$*" >&2; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { err "Missing required command: $1"; exit 1; }
}

detect_os() {
  case "$(uname -s)" in
    Darwin) printf 'darwin'  ;;
    Linux)  printf 'linux'   ;;
    MINGW*|MSYS*|CYGWIN*) printf 'windows' ;;
    *) err "Unsupported OS: $(uname -s)"; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *) err "Unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
}

resolve_version() {
  if [ -n "$VERSION" ]; then
    printf '%s' "$VERSION"
    return
  fi
  # Follow the /releases/latest redirect to discover the tag without using the API.
  local url
  url=$(curl -sIL -o /dev/null -w '%{url_effective}' \
    "https://github.com/${REPO}/releases/latest")
  basename "$url"
}

action_install() {
  require_cmd curl
  require_cmd tar
  require_cmd uname

  local os arch tag version archive_ext archive url tmp
  os=$(detect_os)
  arch=$(detect_arch)
  tag=$(resolve_version)

  if [ -z "$tag" ] || [ "$tag" = "releases" ]; then
    err "Could not resolve a release tag. Set VERSION=v1.0.0 explicitly."
    exit 1
  fi
  version="${tag#v}"

  if [ "$os" = "windows" ]; then
    archive_ext="zip"
    require_cmd unzip
  else
    archive_ext="tar.gz"
  fi

  archive="conjure_${version}_${os}_${arch}.${archive_ext}"
  url="https://github.com/${REPO}/releases/download/${tag}/${archive}"

  tmp=$(mktemp -d)
  trap 'rm -rf "$tmp"' EXIT

  green "  → downloading ${archive} (${tag})"
  if ! curl -fsSL "$url" -o "$tmp/$archive"; then
    err "Download failed: $url"
    exit 1
  fi

  if [ "$archive_ext" = "zip" ]; then
    unzip -q "$tmp/$archive" -d "$tmp"
  else
    tar -xzf "$tmp/$archive" -C "$tmp"
  fi

  local binary="conjure"
  [ "$os" = "windows" ] && binary="conjure.exe"

  if [ ! -f "$tmp/$binary" ]; then
    err "Archive did not contain expected binary: $binary"
    exit 1
  fi

  mkdir -p "$BIN_DIR"
  install -m 0755 "$tmp/$binary" "$CLI_TARGET" 2>/dev/null \
    || { mv "$tmp/$binary" "$CLI_TARGET" && chmod 0755 "$CLI_TARGET"; }
  green "  ✓ Installed $CLI_TARGET (${tag})"

  case ":$PATH:" in
    *":$BIN_DIR:"*) ;;
    *) yellow "  ! $BIN_DIR is not on \$PATH. Add it in your shell config:"
       printf '      export PATH="%s:$PATH"\n' "$BIN_DIR" ;;
  esac

  # Drop the dead symlink from the 0.1.x oh-my-zsh plugin (no longer shipped).
  if [ -L "$OMZ_PLUGIN_FILE" ]; then
    rm "$OMZ_PLUGIN_FILE"
    rmdir "$OMZ_PLUGIN_DIR" 2>/dev/null || true
    yellow "  ! Removed stale oh-my-zsh plugin symlink (replaced by 'conjure shell-init')."
  fi

  printf '\n'
  green "Done. Next steps:"
  printf "  1. Run: conjure setup       (pick provider, store credentials)\n"
  printf "  2. Optional cj() shell helper — append to your shell rc:\n"
  printf '        eval "$(conjure shell-init zsh)"     # or: bash | fish | powershell\n'
}

action_uninstall() {
  if [ -e "$CLI_TARGET" ] || [ -L "$CLI_TARGET" ]; then
    rm "$CLI_TARGET"
    green "  ✓ Removed $CLI_TARGET"
  else
    yellow "  ! $CLI_TARGET not found."
  fi

  if [ -L "$OMZ_PLUGIN_FILE" ]; then
    rm "$OMZ_PLUGIN_FILE"
    rmdir "$OMZ_PLUGIN_DIR" 2>/dev/null || true
    green "  ✓ Removed $OMZ_PLUGIN_FILE"
  fi

  yellow "  ! Stored API key in the keyring is kept. Remove with e.g.:"
  printf '      security delete-generic-password -a "$USER" -s "anthropic-api-key"\n'
}

case "${1:-install}" in
  install)               action_install ;;
  --uninstall|uninstall) action_uninstall ;;
  *) err "Unknown argument: $1"; exit 1 ;;
esac

#!/usr/bin/env bash
#
# conjure installer
#
# Builds and installs the Go binary via `go install ./cmd/conjure`, then
# links it into $BIN_DIR (default ~/.local/bin) so the binary stays on
# $PATH at a stable location while `go install` remains the single source
# of truth for what's actually executed.
#
# Usage:
#   ./install.sh                 # build + install + symlink
#   ./install.sh --uninstall     # remove symlink (binary in $GOBIN is kept)
#   BIN_DIR=~/bin ./install.sh   # custom symlink dir
#

set -uo pipefail

REPO_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
CLI_TARGET="$BIN_DIR/conjure"

# Stale oh-my-zsh plugin from the 0.1.x bash era — cleaned up if present.
OMZ_PLUGIN_DIR="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/conjure"
OMZ_PLUGIN_FILE="$OMZ_PLUGIN_DIR/conjure.plugin.zsh"

green() { printf '\033[32m%s\033[0m\n' "$*"; }
yellow() { printf '\033[33m%s\033[0m\n' "$*"; }
err() { printf '\033[31m%s\033[0m\n' "$*" >&2; }

go_bin_dir() {
  local d
  d=$(go env GOBIN)
  [ -z "$d" ] && d="$(go env GOPATH)/bin"
  printf '%s' "$d"
}

action_install() {
  if ! command -v go >/dev/null 2>&1; then
    err "Go toolchain not found. Install via: brew install go"
    exit 1
  fi

  local gobin
  gobin=$(go_bin_dir)
  green "  → go install ./cmd/conjure  (target: $gobin)"
  ( cd "$REPO_DIR" && go install ./cmd/conjure ) || {
    err "Build failed. If the error mentions a Go version, see go.mod for the required toolchain."
    exit 1
  }
  green "  ✓ Built $gobin/conjure"

  mkdir -p "$BIN_DIR"
  ln -sf "$gobin/conjure" "$CLI_TARGET"
  green "  ✓ Linked $CLI_TARGET → $gobin/conjure"

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
  if [ -L "$CLI_TARGET" ]; then
    rm "$CLI_TARGET"
    green "  ✓ Removed $CLI_TARGET"
  elif [ -e "$CLI_TARGET" ]; then
    yellow "  ! $CLI_TARGET is a regular file (not a symlink); leaving untouched."
    yellow "    Delete manually if it is a leftover from an older install."
  fi

  if [ -L "$OMZ_PLUGIN_FILE" ]; then
    rm "$OMZ_PLUGIN_FILE"
    rmdir "$OMZ_PLUGIN_DIR" 2>/dev/null || true
    green "  ✓ Removed $OMZ_PLUGIN_FILE"
  fi

  if command -v go >/dev/null 2>&1; then
    yellow "  ! Binary kept at $(go_bin_dir)/conjure — remove with:"
    printf '      rm "%s/conjure"\n' "$(go_bin_dir)"
  fi
  yellow "  ! Stored API key in the keyring is kept. Remove with e.g.:"
  printf '      security delete-generic-password -a "$USER" -s "anthropic-api-key"\n'
}

case "${1:-install}" in
  install)   action_install ;;
  --uninstall|uninstall) action_uninstall ;;
  *) err "Unknown argument: $1"; exit 1 ;;
esac
